package main

import (
	"fmt"

	"github.com/houko/nsgo-pose/locales"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// loaded is the catalog res holds, empty until its first load.
func loaded(res *core.Resource[Catalog]) Catalog {
	c, _ := res.Get()
	return c
}

func slotName(t locales.Messages, k Kind) string {
	switch k {
	case KindFace:
		return t.SlotFace()
	case KindCostume:
		return t.SlotCostume()
	default:
		return t.SlotPose()
	}
}

func stepTitle(t locales.Messages, k Kind) string {
	switch k {
	case KindFace:
		return t.StepFace()
	case KindCostume:
		return t.StepCostume()
	default:
		return t.StepPose()
	}
}

// groupLabel names a chip: a gallery's angle, category or hair length in the user's language.
func groupLabel(t locales.Messages, g string) string {
	switch g {
	case "正面":
		return t.AngleFront()
	case "斜め":
		return t.AngleThreeQuarter()
	case "横":
		return t.AngleSide()
	case "上から":
		return t.AngleHigh()
	case "下から":
		return t.AngleLow()
	case "ショート":
		return t.HairShort()
	case "ミディアム":
		return t.HairMedium()
	case "セミロング":
		return t.HairSemiLong()
	case "ロング":
		return t.HairLong()
	case "春":
		return t.GroupSpring()
	case "夏":
		return t.GroupSummer()
	case "秋":
		return t.GroupAutumn()
	case "冬":
		return t.GroupWinter()
	case "パーティ":
		return t.GroupParty()
	case "水着":
		return t.GroupSwimwear()
	}
	return g
}

// itemDetail says what a chosen item is in one line: a face's hair length, an outfit's category and keywords, a pose's name and angle.
func itemDetail(t locales.Messages, it *Item) string {
	if it == nil {
		return t.NotChosen()
	}
	switch it.Kind {
	case KindFace:
		return groupLabel(t, it.Group) + " · " + it.ID
	case KindCostume:
		return groupLabel(t, it.Group) + " · " + it.Title
	default:
		return it.Title + " · " + groupLabel(t, it.Group)
	}
}

// cellAspect is a gallery's thumbnail shape, width over height.
func cellAspect(k Kind) float32 {
	if k == KindFace {
		return 0.8
	}
	return 0.75
}

func cellFit(k Kind) core.ResizeMode {
	if k == KindFace {
		return core.Cover
	}
	return core.Contain
}

type chipOption struct{ value, label string }

// chips is a row of pill-shaped filters that scrolls sideways when it does not fit.
func chips(cx *core.Scope, label, testID string, options []chipOption, value func() string, onChange func(string)) core.Node {
	colors := ui.ColorsOf(cx)
	var pills []core.Node
	for _, o := range options {
		o := o
		on := func() bool { return value() == o.value }
		pills = append(pills, core.View{Role: core.RoleButton, Label: o.label, Selected: on, TestID: testID + "-" + o.label,
			OnPress: func(core.PressEvent) { onChange(o.value) },
			StyleFunc: func() core.Style {
				st := core.Style{Padding: core.Symmetric(7, 16), CornerRadius: 999, BorderWidth: 1, BorderColor: colors.Border(), Background: colors.Surface(), Color: colors.Text()}
				if on() {
					st.Background, st.BorderColor, st.Color = colors.Text(), colors.Text(), colors.Background()
				}
				return st
			},
			Children: []core.Node{core.Text{Content: o.label, Style: core.Style{FontSize: 14}}},
		})
	}
	// A scroll container's minimum height is 0, so a long grid beside it would squeeze it away without one.
	return core.Scroll{Horizontal: true, HideIndicator: true, Style: core.Style{Height: core.Pt(52), MinHeight: core.Pt(52), FlexShrink: 0}, Children: []core.Node{
		core.View{Role: core.RoleGroup, Label: label, Style: core.Style{Direction: core.Row, Gap: 8, Padding: core.Symmetric(8, 0), AlignItems: core.AlignCenter}, Children: pills},
	}}
}

// library is one gallery to choose from: its chips and a grid of thumbnails, columns wide, with the item's name under each when captions is set.
func (s *studio) library(cx *core.Scope, k Kind, columns int, captions bool) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	res := s.catalogs[k]
	filter := s.filter[k]
	chosen := s.chosen[k]
	items := core.NewMemoFunc(cx, func() []Item { return loaded(res).InGroup(filter.Get()) }, nil)
	at := func(i int) (Item, bool) {
		list := items.Get()
		if i < 0 || i >= len(list) {
			return Item{}, false
		}
		return list[i], true
	}
	filters := core.Dynamic(func(cx *core.Scope) core.Node {
		cat := loaded(res)
		if len(cat.Groups) < 2 {
			return nil
		}
		var options []chipOption
		// Poses always show one angle: every angle of every pose at once is the same pose five times.
		if k != KindPose {
			options = append(options, chipOption{"", t.All()})
		}
		for _, g := range cat.Groups {
			options = append(options, chipOption{g, groupLabel(t, g)})
		}
		return chips(cx, t.FiltersLabel(), "chip", options, filter.Get, filter.Set)
	})
	grid := core.List{Style: fill, NumColumns: columns, Gap: 10, ColumnGap: 10, EstimatedSize: 180,
		Count: func() int { return len(items.Get()) },
		Key: func(i int) any {
			it, _ := at(i)
			return it.ID
		},
		Row: func(cx *core.Scope, row core.ListRow) core.Node {
			item := func() Item {
				it, _ := at(row.Index())
				return it
			}
			on := func() bool {
				c := chosen.Get()
				return c != nil && c.ID == item().ID
			}
			label := func() string {
				it := item()
				if it.Title == "" {
					return groupLabel(t, it.Group) + " · " + it.ID
				}
				return it.Title
			}
			children := []core.Node{core.Image{ResizeMode: cellFit(k),
				Style:      core.Style{Width: core.Pct(100), AspectRatio: cellAspect(k), CornerRadius: 10},
				SourceFunc: func() core.ImageSource { return core.Src(item().Thumb) }}}
			if captions {
				children = append(children, core.Text{ContentFunc: label, MaxLines: 1,
					Style: core.Style{FontSize: 13, Padding: core.Symmetric(6, 4)}})
			}
			return core.View{Role: core.RoleButton, LabelFunc: label, Selected: on,
				OnPress: func(core.PressEvent) {
					if it, ok := at(row.Index()); ok {
						s.choose(k, it)
					}
				},
				StyleFunc: func() core.Style {
					st := core.Style{Padding: core.All(4), CornerRadius: 14, BorderWidth: 1, BorderColor: colors.Border(), Background: colors.Surface(), Overflow: core.Hidden}
					if on() {
						st.BorderColor, st.BorderWidth = colors.Accent(), 2.5
					}
					return st
				},
				Children: children,
			}
		},
	}
	state := core.Dynamic(func(cx *core.Scope) core.Node {
		if err := res.Err(); err != nil && len(loaded(res).Items) == 0 {
			return core.View{Style: fill.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Gap: 12, Padding: core.All(24)}), Children: []core.Node{
				core.Text{Content: t.LoadFailed(), Style: ui.TypeHeading},
				core.Text{Content: err.Error(), StyleFunc: func() core.Style {
					return ui.TypeCaption.Merge(core.Style{Color: colors.Muted(), TextAlign: core.TextAlignCenter})
				}},
				ui.Button{Text: t.Retry(), Variant: ui.ButtonSecondary, OnPress: res.Refetch, TestID: "retry"},
			}}
		}
		if len(loaded(res).Items) == 0 {
			return core.View{Style: fill.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter}), Children: []core.Node{
				ui.Spinner{Label: stepTitle(t, k)},
			}}
		}
		return core.View{Style: fill, Children: []core.Node{filters, grid}}
	})
	return core.View{Style: fill, TestID: "library", Children: []core.Node{state}}
}

// thumb is a chosen item's small picture, or an empty well, white where it sits on the inset color.
func thumb(cx *core.Scope, k Kind, it func() *Item, w, h core.Points, round, onInset bool) core.Node {
	colors := ui.ColorsOf(cx)
	radius := core.Points(10)
	if round {
		radius = w / 2
	}
	return core.View{
		StyleFunc: func() core.Style {
			bg := colors.Inset()
			if onInset {
				bg = colors.Surface()
			}
			return core.Style{Width: core.Pt(w), Height: core.Pt(h), CornerRadius: radius, Overflow: core.Hidden, Background: bg,
				AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Color: colors.Muted(), FlexShrink: 0}
		},
		Children: []core.Node{core.Show{
			When: func() bool { return it() != nil },
			Then: func(*core.Scope) core.Node {
				return core.Image{ResizeMode: cellFit(k), Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)},
					SourceFunc: func() core.ImageSource {
						if v := it(); v != nil {
							return core.Src(v.Thumb)
						}
						return core.ImageSource{}
					}}
			},
			Else: func(*core.Scope) core.Node { return core.Icon{SVG: lucide.Plus(), Style: core.Style{FontSize: 18}} },
		}},
	}
}

// lookRow is one line of the look: the item's picture, the gallery's name and what was chosen.
func (s *studio) lookRow(cx *core.Scope, k Kind, chevron bool, onPress func()) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	chosen := s.chosen[k]
	children := []core.Node{
		thumb(cx, k, chosen.Get, 52, 52, k == KindFace, s.wide()),
		core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, Gap: 3}, Children: []core.Node{
			core.Text{Content: slotName(t, k), StyleFunc: func() core.Style {
				return core.Style{FontSize: 13, Color: colors.Muted()}
			}},
			core.Text{ContentFunc: func() string { return itemDetail(t, chosen.Get()) }, MaxLines: 1,
				Style: core.Style{FontSize: 16, FontWeight: core.FontWeightSemibold}},
		}},
	}
	if chevron {
		children = append(children, core.Icon{SVG: lucide.ChevronRight(), StyleFunc: func() core.Style {
			return core.Style{FontSize: 18, Color: colors.Muted()}
		}})
	}
	return core.View{Role: core.RoleButton, TestID: "row-" + kindKey(k),
		LabelFunc: func() string { return slotName(t, k) + ", " + itemDetail(t, chosen.Get()) },
		Selected:  func() bool { return s.wide() && s.step.Get() == k },
		OnPress:   func(core.PressEvent) { onPress() },
		StyleFunc: func() core.Style {
			st := core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 14, Padding: core.Symmetric(12, 14), CornerRadius: 16, Background: colors.Surface(), BorderWidth: 1, BorderColor: colors.Surface()}
			if s.wide() {
				st.Background, st.BorderColor = colors.Inset(), colors.Inset()
				if s.step.Get() == k {
					st.BorderColor = colors.Accent()
				}
			}
			return st
		},
		Children: children,
	}
}

func kindKey(k Kind) string {
	switch k {
	case KindFace:
		return "face"
	case KindCostume:
		return "costume"
	default:
		return "pose"
	}
}

// preview shows the look's references together: the outfit large, or the pose, or the face when nothing else is chosen, with the others small at its corner.
func (s *studio) preview(cx *core.Scope, small core.Points) core.Node {
	colors := ui.ColorsOf(cx)
	order := []Kind{KindCostume, KindPose, KindFace}
	main := func() *Item {
		for _, k := range order {
			if it := s.chosen[k].Get(); it != nil {
				return it
			}
		}
		return nil
	}
	others := core.Dynamic(func(cx *core.Scope) core.Node {
		m := main()
		var nodes []core.Node
		for _, k := range []Kind{KindFace, KindPose, KindCostume} {
			it := s.chosen[k].Get()
			if it == nil || it == m {
				continue
			}
			h := small
			if k != KindFace {
				h = small * 4 / 3
			}
			nodes = append(nodes, core.View{StyleFunc: func() core.Style {
				return core.Style{Padding: core.All(3), CornerRadius: small, Background: colors.Surface()}
			}, Children: []core.Node{thumb(cx, k, s.chosen[k].Get, small, h, k == KindFace, false)}})
		}
		return core.View{Style: core.Style{Position: core.Absolute, Left: core.Pt(16), Bottom: core.Pt(16), Gap: 10}, Children: nodes}
	})
	return core.View{TestID: "preview",
		StyleFunc: func() core.Style {
			return core.Style{FlexGrow: 1, CornerRadius: 24, Overflow: core.Hidden, Background: colors.Inset(), Color: colors.Muted(),
				AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter}
		},
		Children: []core.Node{
			core.Show{
				When: func() bool { return main() != nil },
				Then: func(*core.Scope) core.Node {
					return core.Image{ResizeMode: core.Contain, Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)},
						SourceFunc: func() core.ImageSource {
							if it := main(); it != nil {
								return core.Src(it.Thumb)
							}
							return core.ImageSource{}
						}}
				},
				Else: func(*core.Scope) core.Node { return core.Icon{SVG: lucide.Sparkles(), Style: core.Style{FontSize: 40}} },
			},
			others,
		},
	}
}

// generateButton starts the look, or says what it still lacks.
func (s *studio) generateButton(cx *core.Scope) core.Node {
	t := locales.For(cx)
	return ui.Button{Text: t.Generate(), Icon: lucide.Sparkles, Size: ui.SizeLarge, TestID: "generate",
		Style:    core.Style{Width: core.Pct(100), CornerRadius: 999},
		Disabled: func() bool { return s.request().Validate() != nil },
		Busy:     s.busy.Get, OnPress: s.generate}
}

// needHint says why Generate is off, empty while it is on.
func (s *studio) needHint(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	return core.Show{When: func() bool { return s.request().Validate() != nil }, Then: func(*core.Scope) core.Node {
		return core.Text{Content: t.NeedPose(), TestID: "hint", StyleFunc: func() core.Style {
			return ui.TypeCaption.Merge(core.Style{Color: colors.Muted(), TextAlign: core.TextAlignCenter})
		}}
	}}
}

// countPicker chooses how many pictures one Generate asks for.
func (s *studio) countPicker(cx *core.Scope) core.Node {
	t := locales.For(cx)
	var options []ui.Option[int]
	for _, n := range []int{1, 2, maxCount} {
		options = append(options, ui.Option[int]{Value: n, Label: t.CountOption(n), TestID: fmt.Sprintf("count-%d", n)})
	}
	return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, JustifyContent: core.JustifySpaceBetween, Gap: 12}, Children: []core.Node{
		core.Text{Content: t.CountLabel(), Style: core.Style{FontSize: 15}},
		ui.Tabs[int]{Label: t.CountLabel(), Options: options, Value: s.count.Get, OnChange: s.count.Set, Size: ui.SizeSmall},
	}}
}

func (s *studio) noteInput(cx *core.Scope) core.Node {
	t := locales.For(cx)
	return ui.Input{Value: s.note.Get, OnChange: s.note.Set, Multiline: true, Label: t.NoteLabel(), Placeholder: t.NotePlaceholder(), TestID: "note"}
}

// generatingScreen covers the app while a look is made, dark in either appearance, with a way to stop it.
func (s *studio) generatingScreen(cx *core.Scope) core.Node {
	t := locales.For(cx)
	cx.SetSubtreeAppearance(core.AppearanceDark)
	colors := ui.ColorsOf(cx)
	return core.View{TestID: "generating", Role: core.RoleAlert, Label: t.Generating(),
		StyleFunc: func() core.Style {
			return core.Style{Position: core.Absolute, Top: core.Pt(0), Right: core.Pt(0), Bottom: core.Pt(0), Left: core.Pt(0),
				Background: colors.Background(), Color: colors.Text(), AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Gap: 28}
		},
		Children: []core.Node{
			core.View{StyleFunc: func() core.Style {
				return core.Style{Width: core.Pt(200), Height: core.Pt(268), CornerRadius: 24, Background: colors.Surface(), AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter}
			}, Children: []core.Node{ui.Spinner{Size: 44, Label: t.Generating()}}},
			core.Text{Content: t.Generating(), Style: core.Style{FontSize: 17, FontWeight: core.FontWeightSemibold}},
			ui.Button{Text: t.Cancel(), Variant: ui.ButtonOutline, OnPress: s.cancelGenerating, TestID: "cancel", Style: core.Style{CornerRadius: 999}},
		},
	}
}
