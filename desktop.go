package main

import (
	"github.com/houko/nsgo-pose/locales"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// desktop is the app in a wide window: a top bar of sections over the studio's three columns, the wardrobe, works or settings.
func (s *studio) desktop(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	sections := []ui.Option[int]{
		{Value: sectionStudio, Label: t.TabStudio(), TestID: "section-studio"},
		{Value: sectionWardrobe, Label: t.TabWardrobe(), TestID: "section-wardrobe"},
		{Value: sectionWorks, Label: t.TabWorks(), TestID: "section-works"},
		{Value: sectionSettings, Label: t.TabSettings(), TestID: "section-settings"},
	}
	bar := core.View{
		StyleFunc: func() core.Style {
			return core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 32, Padding: core.Symmetric(14, 24), Background: colors.Surface(),
				Borders: core.Edges{Bottom: 1}, BorderColor: colors.Border()}
		},
		Children: []core.Node{
			brand(cx),
			ui.Tabs[int]{Label: t.TabsLabel(), Options: sections, Value: s.section.Get, OnChange: s.section.Set, TestID: "sections"},
		},
	}
	body := core.Dynamic(func(cx *core.Scope) core.Node {
		switch s.section.Get() {
		case sectionWardrobe:
			return panel(cx, fill.Merge(core.Style{Margin: core.All(24)}), s.library(cx, KindCostume, 6, true))
		case sectionWorks:
			return s.worksTab(cx)
		case sectionSettings:
			return core.View{Style: fill.Merge(core.Style{AlignItems: core.AlignCenter}), Children: []core.Node{
				core.View{Style: fill.Merge(core.Style{Width: core.Pct(100), MaxWidth: core.Pt(680)}), Children: []core.Node{s.settingsPage(cx)}},
			}}
		}
		return s.desktopStudio(cx)
	})
	return core.View{TestID: "desktop", Style: core.Style{FlexGrow: 1, FlexShrink: 1}, Children: []core.Node{bar, body}}
}

// panel is a white rounded card for a column.
func panel(cx *core.Scope, style core.Style, children ...core.Node) core.Node {
	colors := ui.ColorsOf(cx)
	return core.View{
		StyleFunc: func() core.Style {
			return style.Merge(core.Style{Padding: core.All(20), CornerRadius: 24, Background: colors.Surface()})
		},
		Children: children,
	}
}

// desktopStudio is the library on the left, the look or its result in the middle and the look's summary with Generate on the right.
func (s *studio) desktopStudio(cx *core.Scope) core.Node {
	t := locales.For(cx)
	var tabs []ui.Option[Kind]
	var rows []core.Node
	for _, k := range steps {
		k := k
		tabs = append(tabs, ui.Option[Kind]{Value: k, Label: slotName(t, k), TestID: "library-" + kindKey(k)})
		rows = append(rows, s.lookRow(cx, k, false, func() { s.step.Set(k) }))
	}
	left := panel(cx, core.Style{Width: core.Pt(400), FlexShrink: 0, Gap: 8},
		ui.Tabs[Kind]{Label: t.FiltersLabel(), Options: tabs, Value: s.step.Get, OnChange: s.step.Set, Equal: true, TestID: "library-tabs"},
		core.Dynamic(func(cx *core.Scope) core.Node { return s.library(cx, s.step.Get(), 3, s.step.Get() != KindFace) }),
	)
	center := core.View{Style: fill, Children: []core.Node{
		core.Dynamic(func(cx *core.Scope) core.Node {
			if w := s.shown.Get(); w != nil {
				return s.resultCanvas(cx, *w)
			}
			return s.preview(cx, 72)
		}),
	}}
	right := panel(cx, core.Style{Width: core.Pt(360), FlexShrink: 0, Gap: 14},
		core.Text{Content: t.LookTitle(), Role: core.RoleHeader, Style: core.Style{FontSize: 20, FontWeight: core.FontWeightBold}},
		core.View{Style: core.Style{Gap: 10}, Children: rows},
		s.countPicker(cx),
		s.noteInput(cx),
		core.View{Style: core.Style{FlexGrow: 1}},
		s.needHint(cx),
		s.generateButton(cx),
	)
	return core.View{Style: fill.Merge(core.Style{Direction: core.Row, Gap: 24, Padding: core.All(24)}), Children: []core.Node{left, center, right}}
}

// resultCanvas shows a generated picture in the middle column, with save, share and a way back to the look.
func (s *studio) resultCanvas(cx *core.Scope, w work) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	return core.View{TestID: "result",
		StyleFunc: func() core.Style {
			return core.Style{FlexGrow: 1, CornerRadius: 24, Background: colors.Inset(), Padding: core.All(20), Gap: 16}
		},
		Children: []core.Node{
			core.View{Style: core.Style{Direction: core.Row, Gap: 10, JustifyContent: core.JustifyEnd}, Children: []core.Node{
				ui.Button{Text: t.Save(), Icon: lucide.Download, OnPress: func() { s.saveWork(cx, w) }, TestID: "save"},
				ui.Button{Text: t.Share(), Icon: lucide.Share, Variant: ui.ButtonSecondary, OnPress: func() { s.shareWork(cx, w) }, TestID: "share"},
				ui.IconButton{Icon: lucide.X, Label: t.Back(), Variant: ui.IconButtonRaised, OnPress: func() { s.shown.Set(nil) }, TestID: "close-result"},
			}},
			core.Image{Source: core.Src(w.Path), ResizeMode: core.Contain, TestID: "work-image", Label: t.WorkLabel(w.Created),
				Style: fill.Merge(core.Style{Width: core.Pct(100)})},
		},
	}
}
