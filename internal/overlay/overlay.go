package overlay

import (
	"github.com/mattn/go-runewidth"
	"github.com/micro-editor/tcell/v2"
	"github.com/micro-editor/micro/v2/internal/action"
	"github.com/micro-editor/micro/v2/internal/buffer"
	"github.com/micro-editor/micro/v2/internal/config"
	"github.com/micro-editor/micro/v2/internal/display"
	"github.com/micro-editor/micro/v2/internal/screen"
	"github.com/micro-editor/micro/v2/internal/util"
)

type OverlayHandle int
type OverlayFunction func()

type Rect struct {
	X, Y, W, H int
}

var overlay_handle = OverlayHandle(0)
var overlays = make(map[OverlayHandle]OverlayFunction)

// overlay_order preserves creation order so overlays draw deterministically:
// Go map iteration is randomized, which would make overlapping overlays
// flicker depending on which draw happens to run last each frame. Later
// created overlays draw on top.
var overlay_order []OverlayHandle

func DisplayOverlays() {
	// Should an OverlayFunction create or destroy an overlay, that would modify
	// the overlay state while we are iterating through it, so draws are
	// collected first and invoked afterwards.

	draw_fns := make([]OverlayFunction, 0, len(overlay_order))

	for _, h := range overlay_order {
		if draw_fn, ok := overlays[h]; ok {
			draw_fns = append(draw_fns, draw_fn)
		}
	}

	for _, draw_fn := range draw_fns {
		draw_fn()
	}
}

// CreateOverlay creates and registers a new overlay, and returns
// the OverlayHandle associated with it.
func CreateOverlay(draw OverlayFunction) OverlayHandle {
	overlay_handle++
	overlays[overlay_handle] = draw
	overlay_order = append(overlay_order, overlay_handle)
	return overlay_handle
}

// DestroyOverlay destroys/deregisters an existing overlay via its handle.
func DestroyOverlay(overlay OverlayHandle) {
	delete(overlays, overlay)
	for i, h := range overlay_order {
		if h == overlay {
			overlay_order = append(overlay_order[:i], overlay_order[i+1:]...)
			break
		}
	}
}

// DrawRect draws a flat styled rectangle to the provided screen coordinates.
func DrawRect(x, y, w, h int, style tcell.Style) {
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			screen.SetContent(x+xx, y+yy, ' ', nil, style)
		}
	}
}

// DrawText draws styled clipped text to the provided screen coordinates.
func DrawText(text string, x, y, w, h int, style tcell.Style) {
	DrawRect(x, y, w, h, style)

	tabsize := util.IntOpt(config.GlobalSettings["tabsize"])
	text_bytes := []byte(text)
	xx := 0
	yy := 0

	for len(text_bytes) > 0 {
		r, combc, size := util.DecodeCharacter(text_bytes)
		text_bytes = text_bytes[size:]
		width := 0

		switch r {
		case '\t':
			width = tabsize - (xx % tabsize)
		case '\n':
			xx = 0
			yy++
			continue
		default:
			width = runewidth.RuneWidth(r)
		}

		if yy >= h {
			break
		}

		if xx+width <= w {
			screen.SetContent(x+xx, y+yy, r, combc, style)
		}

		xx += width
	}
}

// BufPaneScreenRect returns the bounds of a BufPane in screen coordinates.
func BufPaneScreenRect(bp *action.BufPane) Rect {
	// NOTE: This function is a very thin wrapper around bp.GetView(). As such,
	//       it is maybe a candidate for removal?
	v := bp.GetView()
	return Rect{
		X: v.X,
		Y: v.Y,
		W: v.Width,
		H: v.Height,
	}
}

// BufPaneScreenLoc converts a Loc in the buffer displayed in
// a bufpane to screen coordinates.
func BufPaneScreenLoc(bp *action.BufPane, loc buffer.Loc) buffer.Loc {
	gutter := 0
	bw, ok := bp.BWindow.(*display.BufWindow)
	if ok {
		gutter = bw.GutterOffset()
	}

	v := bp.GetView()
	vloc := bp.VLocFromLoc(loc)
	top := v.StartLine
	yoff := bp.Diff(top, vloc.SLoc)

	return buffer.Loc{
		X: v.X + gutter + vloc.VisualX,
		Y: v.Y + yoff,
	}
}
