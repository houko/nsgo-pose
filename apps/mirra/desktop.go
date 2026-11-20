package main

import (
	"strconv"

	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
	"github.com/yldm-ai/mirra/apps/mirra/locales"
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
				st := core.Style{FlexGrow: 1, FlexBasis: core.Pt(0), Gap: 4, Padding: core.Symmetric(10, 8), CornerRadius: 14, BorderWidth: 1, BorderColor: colors.Border()}
				if on() {
					st.BorderColor, st.Background = colors.Accent(), colors.Inset()
				}
				return st
			},
			Children: []core.Node{
				core.Text{Content: slotEyebrow(t, k), MaxLines: 1, StyleFunc: func() core.Style {
					c := colors.Muted()
					if on() {
						c = colors.Accent()
					}
					return core.Style{FontSize: 9, LetterSpacing: 1, Color: c}
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
	// The library has as many columns as its width holds, about 160 points each.
	libraryColumns := core.NewSignal(cx, 2)
	left := core.View{Style: fill, OnLayout: func(b core.Bounds) {
		if n := max(2, int((b.Size.Width-40)/160)); n != core.Untracked(cx, libraryColumns.Get) {
			libraryColumns.Set(n)
		}
	}, Children: []core.Node{panel(cx, fill.Merge(core.Style{Gap: 10}),
		s.elementTabs(cx),
		core.Dynamic(func(cx *core.Scope) core.Node { return s.library(cx, s.step.Get(), libraryColumns.Get()) }),
	)}}
	center := core.Dynamic(func(cx *core.Scope) core.Node {
		if sh := s.shot.Get(); sh != nil {
			return s.resultCanvas(cx, sh)
		}
		return s.previewCanvas(cx)
	})
	footnote := func() string {
		if s.configured() {
			return t.ViaService(s.serviceName(t))
		}
		return t.ConfigureOnce()
	}
	right := panel(cx, fill.Merge(core.Style{Gap: 12}),
		core.View{Style: core.Style{Gap: 4}, Children: []core.Node{eyebrow(cx, t.EyebrowThisLook), display(t.LookTitle, 22)}},
		core.View{Style: core.Style{Gap: 10}, Children: rows},
		s.styleChips(cx),
		s.noteInput(cx),
		core.View{Style: core.Style{FlexGrow: 1}},
		s.needHint(cx),
		s.generateButton(cx),
		core.Text{ContentFunc: footnote, Style: core.Style{FontSize: 12, TextAlign: core.TextAlignCenter}, StyleFunc: func() core.Style {
			return core.Style{FontSize: 12, TextAlign: core.TextAlignCenter, Color: ui.ColorsOf(cx).Muted()}
		}},
	)
	return s.columns(cx, left, center, right)
}

// The desktop studio's side columns: their widths as shares of the studio's, where the user left them, kept under these keys.
const (
	keyLibraryShare = "column-library"
	keyLookShare    = "column-look"
)

// column is a side column's width: a share of the studio's width, bounded.
type column struct {
	share, min, max float64
	key             string
}

// width is the column's width in a studio w wide, a stored share replacing the default one.
func (c column) width(store core.Storage, w core.Points) core.Points {
	share := c.share
	if v, ok := store.GetItem(c.key); ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 && f < 1 {
			share = f
		}
	}
	return core.Points(min(max(float64(w)*share, c.min), c.max))
}

// columns lays out the library, the canvas and the look side by side. The side columns open at a share of the window, wider on a larger screen, and the user drags the gutters between them to resize them; what they drag to is kept as a share, so it follows the window as it changes.
func (s *studio) columns(cx *core.Scope, left, center, right core.Node) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	store := cx.Storage()
	env := cx.Env()
	const pad, gutter core.Points = 28, 10
	library := column{share: 0.27, min: 300, max: 560, key: keyLibraryShare}
	look := column{share: 0.23, min: 300, max: 460, key: keyLookShare}
	// width is the studio's width, the window's until it is laid out.
	width := core.NewSignal(cx, env.Window().Width)
	// version makes the columns read the stored shares again after a drag.
	version := core.NewSignal(cx, 0)
	size := func(c column) func() core.Points {
		return func() core.Points {
			version.Get()
			return c.width(store, width.Get()) + 2*gutter
		}
	}
	keep := func(c column, size, w core.Points) {
		if w <= 0 {
			return
		}
		if err := store.SetItem(c.key, strconv.FormatFloat(float64(size-2*gutter)/float64(w), 'f', 4, 64)); err != nil {
			s.toastError(t.SaveFailed(), err)
		}
	}
	return ui.ResizablePanelGroup{TestID: "columns", HandleLabel: t.ResizeColumn(),
		Style:    fill.Merge(core.Style{Padding: core.All(pad - gutter)}),
		OnLayout: func(b core.Bounds) { width.Set(b.Size.Width) },
		// The handle sits in the gutter, unseen until the pointer finds it.
		HandleStyleFunc: func() core.Style {
			return core.Style{Background: colors.Accent().Faded(0), Group: "resize-handle",
				GroupStyle: &core.GroupStyle{Name: "resize-handle", Hover: &core.Style{Background: colors.Accent().Faded(0.5)}, Active: &core.Style{Background: colors.Accent()}}}
		},
		OnResize: func(sizes []core.Points) {
			if len(sizes) != 3 {
				return
			}
			w := core.Untracked(cx, width.Get)
			keep(library, sizes[0], w)
			keep(look, sizes[2], w)
			version.Update(func(v int) int { return v + 1 })
		},
		Panels: []base.ResizablePanel{
			{Size: size(library), MinSize: core.Points(library.min) + 2*gutter, MaxSize: core.Points(library.max) + 2*gutter, TestID: "column-library",
				Style: core.Style{Padding: core.Edges{Right: gutter, Left: gutter}}, Children: []core.Node{left}},
			{MinSize: 360, TestID: "column-canvas", Style: core.Style{Padding: core.Edges{Right: gutter, Left: gutter}}, Children: []core.Node{center}},
			{Size: size(look), MinSize: core.Points(look.min) + 2*gutter, MaxSize: core.Points(look.max) + 2*gutter, TestID: "column-look",
				Style: core.Style{Padding: core.Edges{Right: gutter, Left: gutter}}, Children: []core.Node{right}},
		}}
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
			Children: []core.Node{s.thumb(cx, k, s.chosen[k].Get, w, h, k == KindFace)}}
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
		}, func() string { return lookTitle(t, look()) }, func() string { return lookLine(t, look()) }, nil, s.surpriseButton(cx), zoom),
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
		core.Image{Source: s.src(sh.Path), ResizeMode: core.Contain, TestID: "shot-image", Label: lookTitle(t, sh.Req), Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}},
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
