package main

import (
	"github.com/yldm-ai/mirra/locales"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// desktop is the app in a wide window: a top bar over the studio's three columns, the wardrobe or the lookbook, with the AI service's settings in a dialog.
func (s *studio) desktop(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	sections := []ui.Option[int]{
		{Value: sectionStudio, Label: t.TabStudio(), TestID: "section-studio"},
		{Value: sectionWardrobe, Label: t.TabWardrobe(), TestID: "section-wardrobe"},
		{Value: sectionWorks, Label: t.TabWorks(), TestID: "section-works", Content: core.Text{ContentFunc: func() string {
			if n := len(s.works.Get()); n > 0 {
				return t.WorksTabCount(n)
			}
			return t.TabWorks()
		}}},
	}
	brandRow := core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 8}, Children: []core.Node{
		core.Text{Content: t.Brand(), Role: core.RoleHeader, Style: core.Style{FontSize: 24, FontWeight: core.FontWeightBold}},
		caption(cx, t.BrandTagline),
		core.View{StyleFunc: func() core.Style {
			return core.Style{Padding: core.Symmetric(2, 6), CornerRadius: 4, BorderWidth: 1, BorderColor: colors.Accent()}
		}, Children: []core.Node{core.Text{Content: t.OpenSourceBadge(), StyleFunc: func() core.Style {
			return core.Style{FontSize: 9, LetterSpacing: 1.5, Color: colors.Accent()}
		}}}},
	}}
	bar := core.View{
		StyleFunc: func() core.Style {
			return core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 28, Padding: core.Symmetric(12, 28), Borders: core.Edges{Bottom: 1}, BorderColor: colors.Border()}
		},
		Children: []core.Node{
			brandRow,
			ui.Tabs[int]{Label: t.TabsLabel(), Options: sections, Value: s.section.Get, OnChange: s.section.Set, TestID: "sections"},
			core.View{Style: core.Style{FlexGrow: 1}},
			s.statusBadge(cx),
			ui.IconButton{Icon: lucide.SlidersHorizontal, Label: t.AiServiceSettings(), Variant: ui.IconButtonRaised, TestID: "open-settings",
				OnPress: func() { s.settingsShown.Set(true) }},
		},
	}
	body := core.Dynamic(func(cx *core.Scope) core.Node {
		switch s.section.Get() {
		case sectionWardrobe:
			return panel(cx, fill.Merge(core.Style{Margin: core.All(28)}), s.wardrobe(cx, 5))
		case sectionWorks:
			return panel(cx, fill.Merge(core.Style{Margin: core.All(28)}), s.worksPage(cx, 5))
		}
		return s.desktopStudio(cx)
	})
	return core.View{TestID: "desktop", Style: core.Style{FlexGrow: 1, FlexShrink: 1}, Children: []core.Node{bar, body, s.settingsDialog(cx)}}
}

// panel is a rounded column of the desktop.
func panel(cx *core.Scope, style core.Style, children ...core.Node) core.Node {
	colors := ui.ColorsOf(cx)
	return core.View{
		StyleFunc: func() core.Style {
			return style.Merge(core.Style{Padding: core.All(20), CornerRadius: 24, Background: colors.Surface(), BorderWidth: 1, BorderColor: colors.Border()})
		},
		Children: children,
	}
}

// elementTabs choose the gallery the library shows, as the look's three numbered elements.
func (s *studio) elementTabs(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	var tabs []core.Node
	for _, k := range steps {
		on := func() bool { return s.step.Get() == k }
		tabs = append(tabs, core.View{Role: core.RoleTab, Label: slotName(t, k), Selected: on, TestID: "library-" + kindKey(k),
			OnPress: func(core.PressEvent) { s.step.Set(k) },
			StyleFunc: func() core.Style {
				st := core.Style{FlexGrow: 1, FlexBasis: core.Pt(0), Gap: 4, Padding: core.Symmetric(10, 12), CornerRadius: 14, BorderWidth: 1, BorderColor: colors.Border()}
				if on() {
					st.BorderColor, st.Background = colors.Accent(), colors.Inset()
				}
				return st
			},
			Children: []core.Node{
				core.Text{Content: slotEyebrow(t, k), StyleFunc: func() core.Style {
					c := colors.Muted()
					if on() {
						c = colors.Accent()
					}
					return core.Style{FontSize: 9, LetterSpacing: 1.5, Color: c}
				}},
				core.Text{Content: slotName(t, k), Style: core.Style{FontSize: 15, FontWeight: core.FontWeightSemibold}},
			}})
	}
	return core.View{Role: core.RoleTabList, Label: t.FiltersLabel(), Style: core.Style{Direction: core.Row, Gap: 8}, Children: tabs}
}

// desktopStudio is the library on the left, the look or its result in the middle and the look's summary with Generate on the right.
func (s *studio) desktopStudio(cx *core.Scope) core.Node {
	t := locales.For(cx)
	var rows []core.Node
	for _, k := range steps {
		rows = append(rows, s.lookRow(cx, k, func() { s.step.Set(k) }))
	}
	left := panel(cx, core.Style{Width: core.Pt(380), FlexShrink: 0, Gap: 10},
		s.elementTabs(cx),
		core.Dynamic(func(cx *core.Scope) core.Node { return s.library(cx, s.step.Get(), 2) }),
	)
	center := core.Dynamic(func(cx *core.Scope) core.Node {
		if sh := s.shot.Get(); sh != nil {
			return s.resultCanvas(cx, sh)
		}
		return s.previewCanvas(cx)
	})
	footnote := func() string {
		if s.configured() {
			return t.ViaService(serviceName(t, s.service.Get()))
		}
		return t.ConfigureOnce()
	}
	right := panel(cx, core.Style{Width: core.Pt(340), FlexShrink: 0, Gap: 12},
		core.View{Style: core.Style{Gap: 4}, Children: []core.Node{eyebrow(cx, t.EyebrowThisLook), display(t.LookTitle, 22)}},
		core.View{Style: core.Style{Gap: 10}, Children: rows},
		s.noteInput(cx),
		core.View{Style: core.Style{FlexGrow: 1}},
		s.needHint(cx),
		s.generateButton(cx),
		core.Text{ContentFunc: footnote, Style: core.Style{FontSize: 12, TextAlign: core.TextAlignCenter}, StyleFunc: func() core.Style {
			return core.Style{FontSize: 12, TextAlign: core.TextAlignCenter, Color: ui.ColorsOf(cx).Muted()}
		}},
	)
	return core.View{Style: fill.Merge(core.Style{Direction: core.Row, Gap: 20, Padding: core.All(28)}), Children: []core.Node{left, center, right}}
}

// canvasHeader is the canvas's eyebrow and title on the left and its controls on the right.
func canvasHeader(cx *core.Scope, eyebrowText, title, sub func() string, leading core.Node, trailing ...core.Node) core.Node {
	// One line for the title: the controls beside it take the rest of the row.
	texts := core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, Gap: 4}, Children: []core.Node{eyebrow(cx, eyebrowText),
		core.Text{ContentFunc: title, Role: core.RoleHeader, MaxLines: 1, Style: core.Style{FontFamily: displayFont, FontSize: 26, FontWeight: core.FontWeightBold}}}}
	if sub != nil {
		texts.Children = append(texts.Children, caption(cx, sub))
	}
	children := []core.Node{}
	if leading != nil {
		children = append(children, leading)
	}
	children = append(children, texts)
	children = append(children, trailing...)
	return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignStart, Gap: 12}, Children: children}
}

// previewCanvas shows the look's lead picture, zoomable, with the face and pose at its corner to change them.
func (s *studio) previewCanvas(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	look := func() Request { return s.look() }
	zoom := core.View{StyleFunc: func() core.Style {
		return core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 4, Padding: core.All(4), CornerRadius: 999, BorderWidth: 1, BorderColor: colors.Border()}
	}, Children: []core.Node{
		ui.IconButton{Icon: lucide.Minus, Label: t.ZoomOut(), TestID: "zoom-out", Disabled: func() bool { return s.zoom.Get() <= 50 },
			OnPress: func() { s.zoom.Update(func(z int) int { return max(z-25, 50) }) }},
		core.Text{ContentFunc: func() string { return t.ZoomPercent(s.zoom.Get()) }, TestID: "zoom", Style: core.Style{FontSize: 13, Width: core.Pt(44), TextAlign: core.TextAlignCenter}},
		ui.IconButton{Icon: lucide.Plus, Label: t.ZoomIn(), TestID: "zoom-in", Disabled: func() bool { return s.zoom.Get() >= 200 },
			OnPress: func() { s.zoom.Update(func(z int) int { return min(z+25, 200) }) }},
	}}
	quick := func(k Kind, label string, w, h core.Points) core.Node {
		return core.View{Role: core.RoleButton, Label: label, TestID: "quick-" + kindKey(k), OnPress: func(core.PressEvent) { s.step.Set(k) },
			Children: []core.Node{thumb(cx, k, s.chosen[k].Get, w, h, k == KindFace)}}
	}
	stage := core.View{Style: fill.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Overflow: core.Hidden}), Children: []core.Node{
		core.View{StyleFunc: func() core.Style {
			z := float32(s.zoom.Get()) / 100
			return core.Style{Width: core.Pct(70), Height: core.Pct(90), Transform: core.Transform{Scale: z}}
		}, Children: []core.Node{s.heroImage(cx)}},
		core.View{StyleFunc: func() core.Style {
			return core.Style{Position: core.Absolute, Left: core.Pt(0), Bottom: core.Pt(0), Direction: core.Row, Gap: 10, AlignItems: core.AlignEnd,
				Padding: core.All(8), CornerRadius: 18, BorderWidth: 1, BorderColor: colors.Border(), Background: colors.Background().Faded(0.7)}
		}, Children: []core.Node{quick(KindFace, t.ChangeFace(), 56, 56), quick(KindPose, t.ChangePose(), 56, 72)}},
	}}
	return panel(cx, fill.Merge(core.Style{Gap: 16}),
		canvasHeader(cx, func() string {
			if c := s.chosen[KindCostume].Get(); c != nil && c.Group != uploadGroup {
				return t.EyebrowPreviewGroup(groupLabel(t, c.Group))
			}
			return t.EyebrowPreview()
		}, func() string { return lookTitle(t, look()) }, func() string { return lookLine(t, look()) }, nil, zoom),
		stage,
	)
}

// resultCanvas shows a generated picture in the middle column with its actions, the outfits to swap to and Generate again.
func (s *studio) resultCanvas(cx *core.Scope, sh *shot) core.Node {
	t := locales.For(cx)
	back := ui.Button{Text: t.BackToEdit(), Icon: lucide.ChevronLeft, Variant: ui.ButtonOutline, TestID: "close-shot", Style: core.Style{CornerRadius: 999}, OnPress: s.closeShot}
	actions := []core.Node{
		s.favoriteButton(cx, sh),
		ui.IconButton{Icon: lucide.Download, Label: t.Export(), Variant: ui.IconButtonRaised, TestID: "export", OnPress: func() { s.saveFile(cx, sh.Path) }},
		ui.IconButton{Icon: lucide.Trash, Label: t.Delete(), Variant: ui.IconButtonRaised, TestID: "delete", OnPress: s.discardShot},
		core.View{Style: core.Style{Width: core.Pt(150)}, Children: []core.Node{s.saveButton(cx, sh)}},
	}
	stage := core.View{Style: fill.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter}), Children: []core.Node{
		core.Image{Source: core.Src(sh.Path), ResizeMode: core.Contain, TestID: "shot-image", Label: lookTitle(t, sh.Req), Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}},
		core.View{Style: core.Style{Position: core.Absolute, Right: core.Pt(0), Bottom: core.Pt(0)}, Children: []core.Node{s.usedRefs(cx, sh)}},
	}}
	return panel(cx, fill.Merge(core.Style{Gap: 16}),
		canvasHeader(cx, func() string { return shotEyebrow(t, sh) }, func() string { return lookTitle(t, sh.Req) }, nil, back, actions...),
		stage,
		core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignEnd, Gap: 16}, Children: []core.Node{
			core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1}, Children: []core.Node{s.swapRow(cx)}},
			ui.Button{Text: t.RegenerateOnce(), Icon: lucide.RefreshCw, Variant: ui.ButtonOutline, TestID: "regenerate", Style: core.Style{CornerRadius: 999},
				OnPress: func() { s.regenerate(nil) }},
		}},
	)
}
