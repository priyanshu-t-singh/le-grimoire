package render

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/url"
	"strings"
	"time"

	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder

	"le-grimoire/internal/constants"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var WavesharePalette = color.Palette{
	color.Gray{Y: 0},   // 00: Black
	color.Gray{Y: 85},  // 01: Dark Gray
	color.Gray{Y: 170}, // 10: Light Gray
	color.Gray{Y: 255}, // 11: White
}

type Renderer struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
}

// NewRenderer initializes a Renderer based on priority:
// 1. remoteURL (Remote headless container/instance)
// 2. chromePath (Local custom binary)
// 3. Default (Chromedp auto-discovers local Chrome/Chromium)
//
// Both parameters are expected to come from Config (ChromeRemoteURL /
// ChromePath) — pass empty strings to fall through to auto-discovery.
func NewRenderer(baseCtx context.Context, remoteURL string, chromePath string) *Renderer {
	var (
		allocCtx    context.Context
		allocCancel context.CancelFunc
	)

	switch {
	case remoteURL != "":
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(baseCtx, remoteURL)

	case chromePath != "":
		allocCtx, allocCancel = newLocalAllocator(baseCtx, chromePath)

	default:
		allocCtx, allocCancel = newLocalAllocator(baseCtx, "")
	}

	return &Renderer{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
	}
}

func (r *Renderer) Close() {
	if r.allocCancel != nil {
		r.allocCancel()
	}
}

func newLocalAllocator(parent context.Context, execPath string) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.WindowSize(constants.DisplayWidth, constants.DisplayHeight),

		// FIXME:
		// Bypass CORS & Private Network Access restrictions
		// So the image can load, but needs to remove later for security reasons.
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process,BlockInsecurePrivateNetworkRequests"),
	)

	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	return chromedp.NewExecAllocator(parent, opts...)
}

// renders an HTML list view directly into a 2bpp 30000-byte buffer.
func (r *Renderer) RenderListPage(ctx context.Context, htmlContent string) ([]byte, error) {
	taskCtx, cancel := chromedp.NewContext(r.allocCtx)
	defer cancel()

	var pngBuf []byte
	dataURL := "data:text/html," + url.PathEscape(htmlContent)

	err := chromedp.Run(taskCtx,
		chromedp.EmulateViewport(int64(constants.DisplayWidth), int64(constants.DisplayHeight)),
		chromedp.Navigate(dataURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.CaptureScreenshot(&pngBuf),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp list render failed: %w", err)
	}

	return ProcessPNGTo2bpp(pngBuf)
}

const waitForImagesJS = `
Promise.all(
    Array.from(document.images).map(img => {
        if (img.complete) return Promise.resolve();
        return new Promise(resolve => {
            img.addEventListener('load', resolve);
            img.addEventListener('error', resolve); // avoid hanging if an image 404s
        });
    })
)
`

// renders a full HTML book chapter and splits it into discrete widthxheight frames.
func (r *Renderer) RenderBookFrames(ctx context.Context, htmlContent string, lineGap int) ([][]byte, error) {
	taskCtx, cancel := chromedp.NewContext(r.allocCtx)
	defer cancel()

	dataURL := "data:text/html," + url.PathEscape(htmlContent)

	var fullHeight int
	var computedLineHeight float64

	err := chromedp.Run(taskCtx,
		// TODO: Replace constants.DisplayHeight and constants.DisplayWidth with configurable values
		chromedp.EmulateViewport(int64(constants.DisplayWidth), int64(constants.DisplayHeight)),
		chromedp.Navigate(dataURL),
		chromedp.Evaluate(`document.fonts.ready.then(() => true)`, nil),
		chromedp.Evaluate(waitForImagesJS, nil), // <-- Waits for images to download and decode
		chromedp.Evaluate(`Math.ceil(document.body.scrollHeight)`, &fullHeight),
		chromedp.Evaluate(`parseFloat(window.getComputedStyle(document.getElementById('content') || document.body).lineHeight) || 24`, &computedLineHeight),
	)
	if err != nil {
		return nil, fmt.Errorf("measure chapter height: %w", err)
	}

	if fullHeight < constants.DisplayHeight {
		fullHeight = constants.DisplayHeight
	}

	// Set viewport to full height
	err = chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDeviceMetricsOverride(
				int64(constants.DisplayWidth),
				int64(fullHeight),
				1.0,
				false,
			).Do(ctx)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("override full height metrics: %w", err)
	}

	// Compute slice offsets with line-height snapping
	frameHeight := float64(constants.DisplayHeight)
	step := math.Floor(frameHeight/computedLineHeight) * computedLineHeight
	if step <= 0 {
		step = frameHeight * 0.95
	}

	var frames [][]byte
	y := 0.0

	for {
		clipY := y
		// Clamp bottom edge
		if clipY+frameHeight > float64(fullHeight) {
			clipY = float64(fullHeight) - frameHeight
			if clipY < 0 {
				clipY = 0
			}
		}

		var clipPNG []byte
		err := chromedp.Run(taskCtx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				clipPNG, err = page.CaptureScreenshot().
					WithFormat(page.CaptureScreenshotFormatPng).
					WithClip(&page.Viewport{
						X:      0,
						Y:      clipY,
						Width:  float64(constants.DisplayWidth),
						Height: frameHeight,
						Scale:  1.0,
					}).Do(ctx)
				return err
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("capture frame at y=%.1f: %w", clipY, err)
		}

		frame2bpp, err := ProcessPNGTo2bpp(clipPNG)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame2bpp)

		// Stop conditions
		if clipY+frameHeight >= float64(fullHeight) || y+step >= float64(fullHeight) {
			break
		}
		y += step
	}

	return frames, nil
}

// decodes raw image bytes, fits them into widthxheight, dithers, and packs to 2bpp.
func ProcessMangaImage(imgBytes []byte) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("decode manga image: %w", err)
	}

	// Create blank white 400x300 canvas
	canvas := image.NewRGBA(image.Rect(0, 0, constants.DisplayWidth, constants.DisplayHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	// Aspect fit
	srcBounds := src.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	scale := math.Min(float64(constants.DisplayWidth)/srcW, float64(constants.DisplayHeight)/srcH)
	dstW := int(srcW * scale)
	dstH := int(srcH * scale)

	offsetX := (constants.DisplayWidth - dstW) / 2
	offsetY := (constants.DisplayHeight - dstH) / 2

	dstRect := image.Rect(offsetX, offsetY, offsetX+dstW, offsetY+dstH)
	draw.ApproxBiLinear.Scale(canvas, dstRect, src, srcBounds, draw.Over, nil)

	return DitherAndPack(canvas), nil
}

// decodes a PNG and applies Floyd-Steinberg dithering into 2bpp.
func ProcessPNGTo2bpp(pngBuf []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBuf))
	if err != nil {
		return nil, fmt.Errorf("decode png buffer: %w", err)
	}
	return DitherAndPack(img), nil
}

// applies Floyd-Steinberg dithering with WavesharePalette and packs to 2bpp MSB-first.
func DitherAndPack(src image.Image) []byte {
	bounds := image.Rect(0, 0, constants.DisplayWidth, constants.DisplayHeight)
	paletted := image.NewPaletted(bounds, WavesharePalette)
	draw.FloydSteinberg.Draw(paletted, bounds, src, image.Point{})

	out := make([]byte, constants.DisplayWidth*constants.DisplayHeight*2/8) // exactly 30,000 bytes
	bitPos := 0

	for y := 0; y < constants.DisplayHeight; y++ {
		for x := 0; x < constants.DisplayWidth; x++ {
			level := paletted.ColorIndexAt(x, y) // 0..3 (2 bits)
			byteIdx := bitPos / 8
			shift := 6 - (bitPos % 8)
			out[byteIdx] |= level << shift
			bitPos += 2
		}
	}

	return out
}

// produces an emergency 30,000-byte bitmap using stdlib basicfont.
func RenderFallbackErrorBitmap(msg string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, constants.DisplayWidth, constants.DisplayHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	col := color.Black
	point := fixed.Point26_6{X: fixed.I(20), Y: fixed.I(40)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	d.DrawString("ERROR: CONTENT UNAVAILABLE")
	d.Dot = fixed.Point26_6{X: fixed.I(20), Y: fixed.I(70)}

	// Simple word wrap
	words := strings.Fields(msg)
	var line string
	y := 70
	for _, w := range words {
		if len(line)+len(w) > 45 {
			d.Dot = fixed.Point26_6{X: fixed.I(20), Y: fixed.I(y)}
			d.DrawString(line)
			y += 18
			line = w + " "
		} else {
			line += w + " "
		}
	}
	if line != "" {
		d.Dot = fixed.Point26_6{X: fixed.I(20), Y: fixed.I(y)}
		d.DrawString(line)
	}

	d.Dot = fixed.Point26_6{X: fixed.I(20), Y: fixed.I(270)}
	d.DrawString(time.Now().UTC().Format("2006-01-02 15:04:05 UTC") + " - Press E to refresh")

	return DitherAndPack(img)
}
