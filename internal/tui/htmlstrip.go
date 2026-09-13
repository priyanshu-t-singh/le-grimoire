package tui

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// blockElements are HTML tags that produce a paragraph break in plain text.
// Extended to cover the tags Kavita EPUB HTML actually uses.
var blockElements = map[string]bool{
	"p":          true,
	"div":        true,
	"br":         true,
	"li":         true,
	"ol":         true,
	"ul":         true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"blockquote": true,
	"pre":        true,
	"hr":         true,
	"section":    true,
	"article":    true,
	"nav":        true,
	"header":     true,
	"footer":     true,
	"main":       true,
	"figure":     true,
	"figcaption": true,
	"tr":         true,
	"td":         true,
	"th":         true,
}

var noTextElements = map[string]bool{
	"head":     true,
	"script":   true,
	"style":    true,
	"noscript": true,
	"svg":      true,
	"math":     true,
	"iframe":   true,
	"object":   true,
	"canvas":   true,
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`) // basic CSI sequences
var oscRegexp = regexp.MustCompile(`\x1b\][^\x07]*\x07`)       // OSC ... BEL

// removes common bidi control characters that can be used for spoofing.
func removeBiDi(s string) string {
	// Unicode bidi control ranges and specific chars to remove
	bidi := []rune{
		'\u202A', '\u202B', '\u202C', '\u202D', '\u202E',
		'\u2066', '\u2067', '\u2068', '\u2069',
	}
	rb := strings.Builder{}
	for _, r := range s {
		skip := false
		for _, b := range bidi {
			if r == b {
				skip = true
				break
			}
		}
		if !skip {
			rb.WriteRune(r)
		}
	}
	return rb.String()
}

func sanitizeForTerminal(s string, maxLen int) string {
	// Remove explicit CR, remove ANSI/OSC sequences
	s = ansiRegexp.ReplaceAllString(s, "")
	s = oscRegexp.ReplaceAllString(s, "")
	// Remove other ESC-starting sequences loosely (safety)
	s = strings.ReplaceAll(s, "\x1b", "")

	// Strip C0/C1 control characters except LF and TAB
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)

	s = removeBiDi(s)

	// Enforce reasonable max length using runes to avoid breaking UTF-8 characters
	if maxLen > 0 && len(s) > maxLen {
		runes := []rune(s)
		if len(runes) > maxLen {
			s = string(runes[:maxLen]) + "\n\n[output truncated]\n"
		}
	}
	return s
}

// StripHTML converts raw HTML to plain text suitable for terminal display.
//
// It uses the full HTML5 DOM parser (html.Parse) rather than a raw tokenizer
// so that the browser-standard tree is available: structural tags like <head>
// and <body> are correctly separated regardless of what order Kavita delivers
// them
//
// Rules applied during the tree walk:
//   - Subtrees rooted at noTextElements (script, style, svg, …) are skipped.
//   - block-level elements (p, div, h1–h6, li, br, …) emit a newline.
//   - All other element tags are transparent — their text children are kept.
//   - HTML entities are decoded by the parser automatically.
func StripHTML(raw string) string {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		// Fallback: return a best-effort plain text strip via simple tag removal.
		return collapseWhitespace(naiveStripTags(raw))
	}

	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			sb.WriteString(n.Data)

		case html.ElementNode:
			tag := strings.ToLower(n.Data)

			// Silently discard the entire subtree for noTextElements.
			if noTextElements[tag] {
				return
			}

			if blockElements[tag] {
				sb.WriteByte('\n')
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}

			if blockElements[tag] {
				sb.WriteByte('\n')
			}
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return sanitizeForTerminal(collapseWhitespace(sb.String()), 0)
}

// naiveStripTags is a last-resort fallback that removes angle-bracket tags
// with simple string scanning. It handles none of the edge cases that the DOM
// parser does, but it is better than returning raw HTML to the screen.
func naiveStripTags(s string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			sb.WriteRune(r)
		}
	}
	return sanitizeForTerminal(sb.String(), 0)
}

func collapseWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\r", "")

	var sb strings.Builder
	prevNewlines := 0
	for _, r := range s {
		if r == '\n' {
			prevNewlines++
			if prevNewlines <= 2 {
				sb.WriteRune(r)
			}
		} else {
			prevNewlines = 0
			sb.WriteRune(r)
		}
	}

	// Collapse horizontal whitespace: multiple spaces → single space per line.
	lines := strings.Split(sb.String(), "\n")
	for i, line := range lines {
		lines[i] = collapseSpaces(line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func collapseSpaces(s string) string {
	var sb strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) && r != '\n' {
			if !prevSpace {
				sb.WriteRune(' ')
			}
			prevSpace = true
		} else {
			prevSpace = false
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func WordWrap(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 80
	}
	var result []string
	paragraphs := strings.Split(text, "\n")
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(para)
		var line strings.Builder
		for _, word := range words {
			if line.Len() == 0 {
				line.WriteString(word)
			} else if line.Len()+1+len(word) <= maxWidth {
				line.WriteByte(' ')
				line.WriteString(word)
			} else {
				result = append(result, line.String())
				line.Reset()
				line.WriteString(word)
			}
		}
		if line.Len() > 0 {
			result = append(result, line.String())
		}
	}
	return result
}
