package main

import (
	"testing"

	"github.com/nsgo-dev/nsgo/core"
)

// TestZoomStaysOnTheScreen keeps a full-screen picture's zoom between whole and maxZoom, moved no further than its zoom lets it, and toggles it by a double press.
func TestZoomStaysOnTheScreen(t *testing.T) {
	screen := core.Size{Width: 400, Height: 800}
	if z := (zoomState{Scale: 0.5, Offset: core.Point{X: 30}}).clamped(screen); z != (zoomState{Scale: 1}) {
		t.Errorf("zoomed out past whole = %+v, want whole in the middle", z)
	}
	if z := (zoomState{Scale: 9}).clamped(screen); z.Scale != maxZoom {
		t.Errorf("scale = %v, want at most %v", z.Scale, maxZoom)
	}
	// At 2x the picture is 800 by 1600 and moves at most 200 and 400 points each way.
	if z := (zoomState{Scale: 2, Offset: core.Point{X: -999, Y: 999}}).clamped(screen); z.Offset != (core.Point{X: -200, Y: 400}) {
		t.Errorf("offset = %+v, want {-200 400}", z.Offset)
	}
	in := zoomState{Scale: 1}.toggled(core.Point{X: 200, Y: 400}, screen)
	if in != (zoomState{Scale: doubleZoom}) {
		t.Errorf("a double press in the middle = %+v, want %v in the middle", in, doubleZoom)
	}
	// A double press left of the middle brings that side toward it.
	if z := (zoomState{Scale: 1}).toggled(core.Point{X: 100, Y: 400}, screen); z.Offset.X <= 0 || z.Offset.Y != 0 {
		t.Errorf("a double press on the left = %+v, want moved right", z)
	}
	if out := in.toggled(core.Point{}, screen); out != (zoomState{Scale: 1}) {
		t.Errorf("a double press zoomed in = %+v, want whole", out)
	}
}
