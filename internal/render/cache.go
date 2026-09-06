package render

import (
	"fmt"
	"strings"
	"sync"
)

const RenderVersion = 1 // Bump this when HTML/CSS templates change

type FrameCache struct {
	mu     sync.RWMutex
	frames map[string][][]byte // key: "chapter:{id}:page:{page}:v{version}"
}

func NewFrameCache() *FrameCache {
	return &FrameCache{
		frames: make(map[string][][]byte),
	}
}

func (c *FrameCache) key(chapterID string, bookPageIndex int) string {
	return fmt.Sprintf("chapter:%s:page:%d:v%d", chapterID, bookPageIndex, RenderVersion)
}

func (c *FrameCache) chapterPrefix(chapterID string) string {
	return fmt.Sprintf("chapter:%s:", chapterID)
}

// Get returns a single rendered 24-line sub-page frame.
func (c *FrameCache) Get(chapterID string, bookPageIndex int, subPageIndex int) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pageFrames, exists := c.frames[c.key(chapterID, bookPageIndex)]
	if !exists || subPageIndex < 0 || subPageIndex >= len(pageFrames) {
		return nil, false
	}
	return pageFrames[subPageIndex], true
}

func (c *FrameCache) GetAllFrames(chapterID string, bookPageIndex int) ([][]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	frames, exists := c.frames[c.key(chapterID, bookPageIndex)]
	if !exists || len(frames) == 0 {
		return nil, false
	}
	return frames, true
}

func (c *FrameCache) Set(chapterID string, bookPageIndex int, frames [][]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frames[c.key(chapterID, bookPageIndex)] = frames
}

func (c *FrameCache) FrameCount(chapterID string, bookPageIndex int) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.frames[c.key(chapterID, bookPageIndex)])
}

func (c *FrameCache) InvalidatePage(chapterID string, bookPageIndex int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.frames, c.key(chapterID, bookPageIndex))
}

func (c *FrameCache) InvalidateChapter(chapterID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := c.chapterPrefix(chapterID)
	for k := range c.frames {
		if strings.HasPrefix(k, prefix) {
			delete(c.frames, k)
		}
	}
}

func (c *FrameCache) Invalidate(chapterID string) {
	c.InvalidateChapter(chapterID)
}
