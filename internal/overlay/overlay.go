package overlay

import (
	"github.com/mattn/go-runewidth"
	"github.com/micro-editor/micro/v2/internal/action"
	"github.com/micro-editor/micro/v2/internal/buffer"
	"github.com/micro-editor/micro/v2/internal/config"
	"github.com/micro-editor/micro/v2/internal/display"
	"github.com/micro-editor/micro/v2/internal/screen"
	"github.com/micro-editor/micro/v2/internal/util"
	"github.com/micro-editor/tcell/v2"
)

type OverlayHandle int
type OverlayFunction func()

type Rect struct {
	X, Y, W, H int
}

type overlay struct {
	handle OverlayHandle
	draw   OverlayFunction
}

// We use a slice instead of a map so that more than one overlay can be drawn deterministically
var overlays []overlay

var overlay_handle = OverlayHandle(0)

func DisplayOverlays() {
	// Should an OverlayFunction create or destroy an overlay, that would modify
	// the overlays slice while we are iterating through it.
	// For this reason, we copy the overlays slice into temp_overlays.
	temp_overlays := make([]overlay, len(overlays))
	copy(temp_overlays, overlays)

	for _, o := range temp_overlays {
		if live(o.handle) {
			o.draw()
		}
	}
}

// live reports whether an overlay with the given handle is still registered.
// Draw callbacks may destroy overlays mid-frame, so each snapshot entry is
// re-checked before being invoked.
func live(h OverlayHandle) bool {
	for _, o := range overlays {
		if o.handle == h {
			return true
		}
	}
	return false
}

// CreateOverlay creates and registers a new overlay, and returns
// the OverlayHandle associated with it.
func CreateOverlay(draw OverlayFunction) OverlayHandle {
	overlay_handle++
	overlays = append(overlays, overlay{handle: overlay_handle, draw: draw})
	return overlay_handle
}

// DestroyOverlay destroys/deregisters an existing overlay via its handle.
func DestroyOverlay(h OverlayHandle) {
	for i, o := range overlays {
		if o.handle == h {
			overlays = append(overlays[:i], overlays[i+1:]...)
			return
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
