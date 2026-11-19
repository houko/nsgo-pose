package main

import (
	"fmt"

	"github.com/houko/nsgo-pose/locales"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// phone is the app in a narrow window: Studio, Wardrobe, Works and Me over a tab bar.
func (s *studio) phone(cx *core.Scope) core.Node {
	t := locales.For(cx)
	return ui.TabShell{Label: t.TabsLabel(), Selected: s.section.Get, OnChange: s.section.Set, Tabs: []ui.ShellTab{
		{Text: t.TabStudio(), Icon: lucide.House, SystemIcon: "house", Content: s.phoneStudio},
		{Text: t.TabWardrobe(), Icon: lucide.Shirt, SystemIcon: "tshirt", Content: s.phoneWardrobe},
		{Text: t.TabWorks(), Icon: lucide.Images, SystemIcon: "photo.on.rectangle", Content: s.worksTab},
		{Text: t.TabMe(), Icon: lucide.User, SystemIcon: "person", Content: s.settingsPage},
	}}
}

func (s *studio) phoneStudio(cx *core.Scope) core.Node {
	return s.studioNav.Stack(fill, func(cx *core.Scope, r route, _ func() int) base.NavPage {
		switch r.page {
		case pageEditor:
			return s.editorPage(cx)
		case pageWork:
			return s.workPage(cx, s.studioNav, r.work)
		}
		return s.homePage(cx)
	})
}

// brand is Mirra's name at the top of a page, with the tagline beside it in languages that have one.
func brand(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	children := []core.Node{core.Text{Content: t.Brand(), Role: core.RoleHeader, Style: core.Style{FontSize: 28, FontWeight: core.FontWeightBold}}}
	if tag := t.BrandTagline(); tag != "" {
		children = append(children, core.Text{Content: tag, StyleFunc: func() core.Style {
			return core.Style{FontSize: 15, Color: colors.Muted()}
		}})
	}
	return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignBaseline, Gap: 8}, Children: children}
}

func (s *studio) openEditor(k Kind) {
	s.step.Set(k)
	s.studioNav.SetRoutes(route{page: pageHome}, route{page: pageEditor})
}

// firstMissing is the first step still empty, or the pose.
func (s *studio) firstMissing() Kind {
	for _, k := range steps {
		if s.chosen[k].Get() == nil {
			return k
		}
	}
	return KindPose
}

// homePage is the look at a glance: the references together, a row per gallery, and Generate.
func (s *studio) homePage(cx *core.Scope) base.NavPage {
	t := locales.For(cx)
	var rows []core.Node
	for _, k := range steps {
		k := k
		rows = append(rows, s.lookRow(cx, k, true, func() { s.openEditor(k) }))
	}
	return core.Screen{HeaderHidden: true, Title: t.TabStudio(), Style: fill, Children: []core.Node{
		core.Scroll{Style: fill, Children: []core.Node{
			core.View{Style: core.Style{Padding: core.All(20), Gap: 18, PaddingEnv: core.EnvEdges{Top: core.EnvSafeArea}}, Children: []core.Node{
				brand(cx),
				core.View{Role: core.RoleButton, Label: stepTitle(t, KindFace), OnPress: func(core.PressEvent) { s.openEditor(s.firstMissing()) },
					Style: core.Style{Height: core.Pt(300)}, Children: []core.Node{s.preview(cx, 52)}},
				core.Text{Content: t.LookTitle(), Role: core.RoleHeader, Style: core.Style{FontSize: 20, FontWeight: core.FontWeightBold}},
				core.View{Style: core.Style{Gap: 10}, Children: rows},
				s.countPicker(cx),
				s.noteInput(cx),
			}},
		}},
		core.View{Style: core.Style{Padding: core.Symmetric(12, 20), Gap: 8}, Children: []core.Node{
			s.needHint(cx),
			s.generateButton(cx),
		}},
	}}
}

// editorPage puts the look together a gallery at a time: face, outfit, then pose.
func (s *studio) editorPage(cx *core.Scope) base.NavPage {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	var options []ui.Option[Kind]
	for i, k := range steps {
		k, n := k, i+1
		options = append(options, ui.Option[Kind]{Value: k, TestID: "step-" + kindKey(k),
			Content: core.Text{ContentFunc: func() string {
				label := t.StepTab(fmt.Sprint(n), slotName(t, k))
				if s.chosen[k].Get() != nil && s.step.Get() != k {
					label += " ✓"
				}
				return label
			}}, Label: t.StepTab(fmt.Sprint(n), slotName(t, k))})
	}
	bottom := core.Dynamic(func(cx *core.Scope) core.Node {
		switch s.step.Get() {
		case KindFace:
			return ui.Button{Text: t.NextCostume(), Size: ui.SizeLarge, TestID: "next", Style: core.Style{Width: core.Pct(100), CornerRadius: 999},
				Color: colors.Text, OnPress: func() { s.step.Set(KindCostume) }}
		case KindCostume:
			return ui.Button{Text: t.NextPose(), Size: ui.SizeLarge, TestID: "next", Style: core.Style{Width: core.Pct(100), CornerRadius: 999},
				Color: colors.Text, OnPress: func() { s.step.Set(KindPose) }}
		}
		return core.View{Style: core.Style{Gap: 8}, Children: []core.Node{s.needHint(cx), s.generateButton(cx)}}
	})
	return ui.NavPage(cx, core.Screen{TitleFunc: func() string { return stepTitle(t, s.step.Get()) }, Style: fill, Children: []core.Node{
		core.View{Style: fill.Merge(core.Style{Padding: core.Symmetric(8, 20), Gap: 12}), Children: []core.Node{
			ui.Tabs[Kind]{Label: t.StepsLabel(), Options: options, Value: s.step.Get, OnChange: s.step.Set, Equal: true, TestID: "steps"},
			core.Dynamic(func(cx *core.Scope) core.Node { return s.chosenCard(cx, s.step.Get()) }),
			core.Dynamic(func(cx *core.Scope) core.Node {
				k := s.step.Get()
				columns := 3
				if k == KindFace {
					columns = 4
				}
				return s.library(cx, k, columns, k != KindFace)
			}),
		}},
		core.View{Style: core.Style{Padding: core.Symmetric(12, 20)}, Children: []core.Node{bottom}},
	}}, ui.NavHeader{OnBack: func() { s.studioNav.Pop(1) }, BackLabel: t.Back(),
		Trailing: []core.Node{ui.Button{Text: t.Done(), Variant: ui.ButtonGhost, TestID: "done", OnPress: func() { s.studioNav.Pop(1) }}}})
}

// chosenCard shows what a step has chosen so far, above its gallery.
func (s *studio) chosenCard(cx *core.Scope, k Kind) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	chosen := s.chosen[k]
	top := func() string {
		// A face's detail already ends with its ID.
		if it := chosen.Get(); it != nil && k != KindFace {
			return t.SelectedItem(it.ID)
		}
		return slotName(t, k)
	}
	return core.View{TestID: "chosen",
		StyleFunc: func() core.Style {
			return core.Style{Direction: core.Row, Gap: 16, Padding: core.All(12), CornerRadius: 20, Background: colors.Surface(), AlignItems: core.AlignCenter}
		},
		Children: []core.Node{
			thumb(cx, k, chosen.Get, 84, 112, false, false),
			core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, Gap: 6}, Children: []core.Node{
				core.Text{ContentFunc: top, StyleFunc: func() core.Style { return core.Style{FontSize: 13, Color: colors.Muted()} }},
				core.Text{ContentFunc: func() string { return itemDetail(t, chosen.Get()) }, MaxLines: 3,
					Style: core.Style{FontSize: 17, FontWeight: core.FontWeightSemibold}},
			}},
		},
	}
}

// phoneWardrobe browses the outfits; choosing one puts it in the look.
func (s *studio) phoneWardrobe(cx *core.Scope) core.Node {
	t := locales.For(cx)
	return ui.NavPage(cx, core.Screen{Title: t.TabWardrobe(), Style: fill, Children: []core.Node{
		core.View{Style: fill.Merge(core.Style{Padding: core.Symmetric(4, 20)}), Children: []core.Node{s.library(cx, KindCostume, 2, true)}},
	}}, ui.NavHeader{})
}
