package main

import (
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"

	"github.com/yldm-ai/mirra/apps/mirra/locales"
)

// swipeDistance is how far a horizontal swipe over a picture goes to step to the next work or back.
const swipeDistance = 60

// shownWorks are the works the Works page lists, all of them or the favorites, in its order: the ones a work on show steps through.
func (s *studio) shownWorks() []work {
	ws := s.works.Get()
	if !s.favoritesOn.Get() {
		return ws
	}
	var out []work
	for _, w := range ws {
		if w.Favorite {
			out = append(out, w)
		}
	}
	return out
}

// workIndex is where the picture on show is among the shown works, -1 when it is not one of them.
func (s *studio) workIndex(sh *shot) int {
	if sh == nil || sh.Saved == nil {
		return -1
	}
	for i, w := range s.shownWorks() {
		if w.Path == sh.Saved.Path {
			return i
		}
	}
	return -1
}

// canStep says whether there is a shown work d places from the one on show.
func (s *studio) canStep(d int) bool {
	i := s.workIndex(s.shot.Get())
	return i >= 0 && i+d >= 0 && i+d < len(s.shownWorks())
}

// stepWork shows the work d places from the one on show, as opening it from Works does, in the same place.
func (s *studio) stepWork(d int) {
	if !s.canStep(d) {
		return
	}
	ws := s.shownWorks()
	s.shot.Set(s.workShot(ws[s.workIndex(s.shot.Get())+d]))
}

// stepKeys binds the left and right arrows to the previous and next work while what cx shows is on screen.
func (s *studio) stepKeys(cx *core.Scope) {
	cx.OnKey("left", func() { s.stepWork(-1) })
	cx.OnKey("right", func() { s.stepWork(1) })
}

// swipeToStep is a horizontal swipe over a picture stepping to the next work (leftwards) or the previous one.
func (s *studio) swipeToStep(e core.PanEvent) {
	if e.Phase != core.GestureEnded {
		return
	}
	switch {
	case e.Translation.X <= -swipeDistance:
		s.stepWork(1)
	case e.Translation.X >= swipeDistance:
		s.stepWork(-1)
	}
}

// swipeActivation lets a pan begin only as a sideways drag, leaving vertical ones to the scroll views around it.
var swipeActivation = core.PanActivation{ActiveOffsetX: 16, FailOffsetY: 24}

// workPosition is "3 / 12" for the third of twelve shown works, "" when the picture on show is not one of them.
func (s *studio) workPosition() string {
	i := s.workIndex(s.shot.Get())
	if i < 0 {
		return ""
	}
	return locales.For(s.cx).WorkPosition(i+1, len(s.shownWorks()))
}

// stepButtons are the previous and next arrows over a picture's sides, each shown while there is a work that way.
func (s *studio) stepButtons(cx *core.Scope) []core.Node {
	t := locales.For(cx)
	arrow := func(d int, icon func() *core.SVG, label, id string, side core.Style) core.Node {
		return core.Show{When: func() bool { return s.canStep(d) }, Then: func(cx *core.Scope) core.Node {
			st := core.Style{Position: core.Absolute, Top: core.Pct(50), Margin: core.Edges{Top: -22}}
			return core.View{Style: st.Merge(side), Children: []core.Node{
				ui.IconButton{Icon: icon, Label: label, Variant: ui.IconButtonRaised, TestID: id, OnPress: func() { s.stepWork(d) }},
			}}
		}}
	}
	return []core.Node{
		arrow(-1, lucide.ChevronLeft, t.PreviousWork(), "previous-work", core.Style{Left: core.Pt(12)}),
		arrow(1, lucide.ChevronRight, t.NextWork(), "next-work", core.Style{Right: core.Pt(12)}),
	}
}

// fullScreenButton opens the picture on show over the whole window.
func (s *studio) fullScreenButton(cx *core.Scope) core.Node {
	return ui.IconButton{Icon: lucide.Maximize2, Label: locales.For(cx).ViewFullScreen(), Variant: ui.IconButtonRaised, TestID: "full-screen", OnPress: func() { s.fullScreen.Set(true) }}
}

// viewer is the picture on show over the whole window, on black, with the arrows to the other works: Escape, back or × closes it.
func (s *studio) viewer(cx *core.Scope) core.Node {
	t := locales.For(cx)
	s.stepKeys(cx)
	closeViewer := func() { s.fullScreen.Set(false) }
	white := core.RGB(0xffffff)
	picture := core.Dynamic(func(cx *core.Scope) core.Node {
		sh := s.shot.Get()
		if sh == nil {
			return nil
		}
		return core.Image{Source: s.src(sh.Path), ResizeMode: core.Contain, TestID: "full-screen-image", Label: lookTitle(t, sh.Req),
			Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}}
	})
	children := []core.Node{
		core.View{Style: fill, OnPan: s.swipeToStep, PanActivation: swipeActivation, Children: []core.Node{picture}},
		core.View{Style: core.Style{Position: core.Absolute, Top: core.Pt(16), Left: core.Pt(16), Right: core.Pt(16), Direction: core.Row, AlignItems: core.AlignCenter},
			Children: []core.Node{
				core.Text{ContentFunc: s.workPosition, TestID: "work-position", Style: core.Style{FontSize: 14, Color: white, FlexGrow: 1}},
				ui.IconButton{Icon: lucide.X, Label: t.Close(), Variant: ui.IconButtonRaised, TestID: "close-full-screen", OnPress: closeViewer},
			}},
	}
	children = append(children, s.stepButtons(cx)...)
	return core.Modal{Animation: core.ModalFade, TestID: "full-screen", Label: t.ViewFullScreen(),
		OnRequestClose: func(string) { closeViewer() },
		StatusBar:      core.StatusBarSetting{Hidden: core.ToggleOn},
		Style: core.Style{FlexGrow: 1, Background: core.RGB(0x000000), PaddingEnv: core.EnvEdges{Top: core.EnvSafeArea, Bottom: core.EnvSafeArea}},
		Children: children,
	}
}

// pictureButton is the picture on show, filling its place: pressing it opens it full screen, and a sideways swipe steps to the next or previous work.
func (s *studio) pictureButton(cx *core.Scope, sh *shot) core.Node {
	t := locales.For(cx)
	return core.View{Role: core.RoleButton, Label: t.ViewFullScreen(), TestID: "open-full-screen", Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)},
		OnPress: func(core.PressEvent) { s.fullScreen.Set(true) }, OnPan: s.swipeToStep, PanActivation: swipeActivation,
		Children: []core.Node{
			core.Image{Source: s.src(sh.Path), ResizeMode: core.Contain, TestID: "shot-image", Label: lookTitle(t, sh.Req), Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}},
		}}
}
