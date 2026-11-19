// Command nsgo-pose dresses a pose, or the user's own photo, in an outfit and gives it a face, taking the three from public reference galleries and combining them with an OpenAI-compatible image model.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/houko/nsgo-pose/locales"
	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func main() {
	app := App{Client: &http.Client{Timeout: 5 * time.Minute}, Sites: DefaultSites}
	nsgo.Run(app.Root, nsgo.WindowSize(440, 880))
}

// App is the app with what it reaches outside: tests give it local servers and a temporary directory.
type App struct {
	Client *http.Client
	Sites  Sites
	// WorksDir keeps generated pictures; the app's data directory when empty.
	WorksDir string
}

type baseMode int

const (
	basePose baseMode = iota
	basePhoto
)

type page int

const (
	pageHome page = iota
	pagePicker
	pageWork
)

type route struct {
	page page
	kind Kind
	work work
}

const apiKeyAccount = "api-key"

// studio is the app's state, shared by its tabs.
type studio struct {
	App
	cx      *core.Scope
	toaster *ui.Toaster
	tab     *core.Signal[int]

	mode                *core.Signal[baseMode]
	pose, costume, face *core.Signal[*Item]
	photo               *core.Signal[*core.PickedFile]
	note                *core.Signal[string]
	busy                *core.Signal[bool]

	baseURL, apiKey, model, size *core.Signal[string]

	works    *core.Signal[[]work]
	wanted   [3]*core.Signal[bool]
	catalogs [3]*core.Resource[Catalog]

	createNav, worksNav *base.Nav[route]
}

var fill = core.Style{FlexGrow: 1, FlexShrink: 1}

// Root is the app: creating, the works made so far, and settings, over a tab bar.
func (a App) Root(cx *core.Scope) core.Node {
	s := newStudio(cx, a)
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	return core.View{
		StyleFunc: func() core.Style {
			return fill.Merge(core.Style{Background: colors.Background(), Color: colors.Text()})
		},
		Children: []core.Node{
			ui.TabShell{Label: t.TabsLabel(), Selected: s.tab.Get, OnChange: s.tab.Set, Tabs: []ui.ShellTab{
				{Text: t.TabCreate(), Icon: lucide.Sparkles, SystemIcon: "sparkles", Content: s.createTab},
				{Text: t.TabWorks(), Icon: lucide.Images, SystemIcon: "photo.on.rectangle", Content: s.worksTab},
				{Text: t.TabSettings(), Icon: lucide.Settings, SystemIcon: "gearshape", Content: s.settingsTab},
			}},
			ui.Toasts{Toaster: s.toaster, Placement: ui.ToastTop},
		},
	}
}

func newStudio(cx *core.Scope, a App) *studio {
	if a.Client == nil {
		a.Client = http.DefaultClient
	}
	store := cx.Storage()
	stored := func(key, fallback string) *core.Signal[string] {
		v, ok := store.GetItem(key)
		if !ok {
			v = fallback
		}
		return core.NewSignal(cx, v)
	}
	s := &studio{
		App:     a,
		cx:      cx,
		toaster: ui.NewToaster(cx),
		tab:     core.NewSignal(cx, 0),
		mode:    core.NewSignal(cx, basePose),
		pose:    core.NewSignal[*Item](cx, nil),
		costume: core.NewSignal[*Item](cx, nil),
		face:    core.NewSignal[*Item](cx, nil),
		photo:   core.NewSignal[*core.PickedFile](cx, nil),
		note:    core.NewSignal(cx, ""),
		busy:    core.NewSignal(cx, false),
		baseURL: stored("baseURL", defaultBaseURL),
		apiKey:  core.NewSignal(cx, ""),
		model:   stored("model", defaultModel),
		size:    stored("size", defaultSize),
		works:   core.NewSignalFunc[[]work](cx, nil, nil),
	}
	s.createNav = base.NewNav(cx, route{page: pageHome})
	s.worksNav = base.NewNav(cx, route{page: pageHome})
	for k := range s.catalogs {
		kind := Kind(k)
		s.wanted[k] = core.NewSignal(cx, false)
		s.catalogs[k] = core.NewResource(cx, func() (Kind, bool) { return kind, s.wanted[kind].Get() },
			func(ctx context.Context, k Kind, _ core.ResourceInfo) (Catalog, error) {
				return FetchCatalog(ctx, s.Client, s.Sites, k)
			})
	}

	secrets := cx.Secrets()
	dir := s.worksDir()
	go func() {
		key, ok, keyErr := secrets.Get(apiKeyAccount)
		list, listErr := listWorks(dir)
		cx.Post(func() {
			if keyErr == nil && ok && s.apiKey.Get() == "" {
				s.apiKey.Set(key)
			}
			if listErr == nil {
				s.works.Set(list)
			}
		})
	}()
	return s
}

func (s *studio) worksDir() string {
	if s.WorksDir != "" {
		return s.WorksDir
	}
	if d := s.cx.DataDir(); d != "" {
		return filepath.Join(d, "works")
	}
	return filepath.Join(os.TempDir(), "nsgo-pose", "works")
}

func (s *studio) slot(k Kind) *core.Signal[*Item] {
	switch k {
	case KindPose:
		return s.pose
	case KindCostume:
		return s.costume
	default:
		return s.face
	}
}

// request is what the user has put together; reading it inside a binding follows every choice.
func (s *studio) request() Request {
	r := Request{Face: s.face.Get(), Note: s.note.Get()}
	if s.mode.Get() == basePose {
		r.Base.Item = s.pose.Get()
		r.Costume = s.costume.Get()
	} else if p := s.photo.Get(); p != nil {
		r.Base = Source{Photo: p.Data, Name: p.Name, MIME: p.MIME}
	}
	return r
}

func (s *studio) toastError(title string, err error) {
	s.toaster.Show(ui.ToastContent{Type: ui.ToastError, Title: title, Description: err.Error()})
}

func (s *studio) generate() {
	t := locales.For(s.cx)
	req := s.request()
	if req.Validate() != nil {
		return
	}
	cfg := Config{BaseURL: s.baseURL.Get(), APIKey: s.apiKey.Get(), Model: s.model.Get(), Size: s.size.Get()}
	if strings.TrimSpace(cfg.APIKey) == "" {
		s.toaster.Show(ui.ToastContent{Type: ui.ToastWarning, Title: t.NeedKey()})
		s.tab.Set(2)
		return
	}
	s.busy.Set(true)
	dir, now, client := s.worksDir(), s.cx.Now(), s.Client
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		data, err := Generate(ctx, client, cfg, req)
		var w work
		if err == nil {
			w, err = saveWork(dir, data, now)
		}
		s.cx.Post(func() {
			s.busy.Set(false)
			if err != nil {
				s.toastError(t.GenerateFailed(), err)
				return
			}
			s.works.Update(func(ws []work) []work { return append([]work{w}, ws...) })
			s.createNav.Push(route{page: pageWork, work: w})
		})
	}()
}

// ---- Create ----

func (s *studio) createTab(cx *core.Scope) core.Node {
	return s.createNav.Stack(fill, func(cx *core.Scope, r route, _ func() int) base.NavPage {
		switch r.page {
		case pagePicker:
			return s.pickerPage(cx, r.kind)
		case pageWork:
			return s.workPage(cx, s.createNav, r.work)
		}
		return s.homePage(cx)
	})
}

func (s *studio) openPicker(k Kind) {
	s.wanted[k].Set(true)
	s.createNav.Push(route{page: pagePicker, kind: k})
}

func (s *studio) pickPhoto() {
	s.cx.PickImages(core.PickOptions{Limit: 1, MaxDimension: 2048}, func(files []core.PickedFile) {
		if len(files) == 1 {
			f := files[0]
			s.photo.Set(&f)
		}
	})
}

func itemSource(sig *core.Signal[*Item]) func() core.ImageSource {
	return func() core.ImageSource {
		if it := sig.Get(); it != nil {
			return core.Src(it.Thumb)
		}
		return core.ImageSource{}
	}
}

func (s *studio) homePage(cx *core.Scope) base.NavPage {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	base := core.Dynamic(func(cx *core.Scope) core.Node {
		if s.mode.Get() == basePose {
			return slotCard{Title: t.SlotPose(), Image: itemSource(s.pose), Fit: core.Contain, TestID: "slot-pose",
				OnPress: func() { s.openPicker(KindPose) }}
		}
		return slotCard{Title: t.SlotPhoto(), TestID: "slot-photo", Fit: core.Cover,
			Image: func() core.ImageSource {
				if p := s.photo.Get(); p != nil {
					return core.ImageSource{Data: p.Data, Size: p.Size}
				}
				return core.ImageSource{}
			},
			OnPress: s.pickPhoto, OnClear: func() { s.photo.Set(nil) }}
	})
	hint := func() string {
		switch s.request().Validate() {
		case ErrNoBase:
			if s.mode.Get() == basePhoto {
				return t.NeedPhoto()
			}
			return t.NeedPose()
		case ErrNothingToDo:
			return t.NeedPhotoChange()
		}
		return ""
	}
	return ui.NavPage(cx, core.Screen{Title: t.TabCreate(), Children: []core.Node{
		core.Scroll{Style: fill, Children: []core.Node{
			core.View{Style: core.Style{Padding: core.All(16), Gap: 16}, Children: []core.Node{
				ui.Tabs[baseMode]{Label: t.BaseLabel(), Equal: true, Value: s.mode.Get, OnChange: s.mode.Set, TestID: "base-mode",
					Options: []ui.Option[baseMode]{{Value: basePose, Label: t.BasePose(), TestID: "mode-pose"}, {Value: basePhoto, Label: t.BasePhoto(), TestID: "mode-photo"}}},
				core.View{Style: core.Style{Direction: core.Row, Gap: 10}, Children: []core.Node{
					base,
					// Outfits go on poses only: the outfit gallery's terms forbid depicting real people with its images.
					core.Show{When: func() bool { return s.mode.Get() == basePose }, Then: func(*core.Scope) core.Node {
						return slotCard{Title: t.SlotCostume(), Image: itemSource(s.costume), Fit: core.Contain, TestID: "slot-costume",
							OnPress: func() { s.openPicker(KindCostume) }, OnClear: func() { s.costume.Set(nil) }}
					}},
					slotCard{Title: t.SlotFace(), Image: itemSource(s.face), Fit: core.Cover, TestID: "slot-face",
						OnPress: func() { s.openPicker(KindFace) }, OnClear: func() { s.face.Set(nil) }},
				}},
				ui.Input{Value: s.note.Get, OnChange: s.note.Set, Multiline: true, Label: t.NoteLabel(), Placeholder: t.NotePlaceholder(), TestID: "note"},
				core.Show{When: func() bool { return hint() != "" }, Then: func(*core.Scope) core.Node {
					return core.Text{ContentFunc: hint, TestID: "hint", StyleFunc: func() core.Style {
						return ui.TypeCaption.Merge(core.Style{Color: colors.Muted(), TextAlign: core.TextAlignCenter})
					}}
				}},
				ui.Button{Text: t.Generate(), Icon: lucide.Sparkles, Size: ui.SizeLarge, Style: core.Style{Width: core.Pct(100)}, TestID: "generate",
					Disabled: func() bool { return s.request().Validate() != nil }, Busy: s.busy.Get, OnPress: s.generate},
			}},
		}},
	}}, ui.NavHeader{})
}

// slotCard is one of the three choices: a 3:4 well showing the chosen image, or a plus while empty, with its name below.
type slotCard struct {
	Title   string
	Image   func() core.ImageSource
	Fit     core.ResizeMode
	OnPress func()
	// OnClear, when set, adds a button that empties the slot.
	OnClear func()
	TestID  string
}

func (c slotCard) Render(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	has := func() bool {
		src := c.Image()
		return src.URI != "" || len(src.Data) > 0
	}
	well := core.View{TestID: c.TestID, Role: core.RoleButton, OnPress: func(core.PressEvent) { c.OnPress() },
		LabelFunc: func() string {
			if has() {
				return c.Title
			}
			return t.ChooseSlot(c.Title)
		},
		StyleFunc: func() core.Style {
			return core.Style{Width: core.Pct(100), AspectRatio: 0.75, CornerRadius: 12, Overflow: core.Hidden,
				AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, BorderWidth: 1,
				BorderColor: colors.Border(), Background: colors.Surface(), Color: colors.Muted()}
		},
		Children: []core.Node{core.Show{When: has,
			Then: func(*core.Scope) core.Node {
				return core.Image{SourceFunc: c.Image, ResizeMode: c.Fit, Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}}
			},
			Else: func(*core.Scope) core.Node { return core.Icon{SVG: lucide.Plus(), Style: core.Style{FontSize: 28}} },
		}},
	}
	children := []core.Node{well}
	if c.OnClear != nil {
		children = append(children, core.Show{When: has, Then: func(*core.Scope) core.Node {
			return core.View{Style: core.Style{Position: core.Absolute, Top: core.Pt(4), Right: core.Pt(4)}, Children: []core.Node{
				ui.IconButton{Icon: lucide.X, Label: t.RemoveSlot(c.Title), Variant: ui.IconButtonRaised, Size: 28, OnPress: c.OnClear, TestID: c.TestID + "-clear"},
			}}
		}})
	}
	return core.View{Style: core.Style{FlexGrow: 1, FlexBasis: core.Pt(0), Gap: 6}, Children: []core.Node{
		core.View{Children: children},
		core.Text{Content: c.Title, Style: ui.TypeCaption.Merge(core.Style{TextAlign: core.TextAlignCenter})},
	}}
}

// ---- Picker ----

func groupLabel(t locales.Messages, g string) string {
	switch g {
	case "全身1人・標準系":
		return t.GroupPoseStandard()
	case "全身1人・セクシー系":
		return t.GroupPoseGlamour()
	case "上半身1人":
		return t.GroupPoseUpper()
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
	case "face":
		return t.GroupFace()
	case "styling":
		return t.GroupStyling()
	}
	return g
}

func (s *studio) pickerPage(cx *core.Scope, kind Kind) base.NavPage {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	res := s.catalogs[kind]
	chosen := s.slot(kind)
	group := core.NewSignal(cx, "")
	items := core.NewMemoFunc(cx, func() []Item { return loaded(res).InGroup(group.Get()) }, nil)
	title := map[Kind]string{KindPose: t.PickPose(), KindCostume: t.PickCostume(), KindFace: t.PickFace()}[kind]
	fit := core.Contain
	if kind == KindFace {
		fit = core.Cover
	}
	at := func(i int) (Item, bool) {
		list := items.Get()
		if i < 0 || i >= len(list) {
			return Item{}, false
		}
		return list[i], true
	}

	groups := core.Dynamic(func(cx *core.Scope) core.Node {
		cat := loaded(res)
		if len(cat.Groups) < 2 {
			return nil
		}
		options := []ui.Option[string]{{Value: "", Label: t.GroupAll()}}
		for _, g := range cat.Groups {
			options = append(options, ui.Option[string]{Value: g, Label: groupLabel(t, g)})
		}
		// A scroll container's minimum height is 0, so a long list beside it would squeeze it away without one.
		return core.Scroll{Horizontal: true, HideIndicator: true, Style: core.Style{Height: core.Pt(48), MinHeight: core.Pt(48)}, Children: []core.Node{
			core.View{Style: core.Style{Padding: core.Symmetric(8, 16)}, Children: []core.Node{
				ui.Tabs[string]{Label: t.GroupsLabel(), Options: options, Value: group.Get, OnChange: group.Set, TestID: "groups"},
			}},
		}}
	})

	grid := core.List{Style: fill.Merge(core.Style{FlexBasis: core.Pt(0), Padding: core.All(16)}), NumColumns: 3, Gap: 8, ColumnGap: 8, EstimatedSize: 150,
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
			return core.View{Role: core.RoleButton,
				LabelFunc: func() string { return item().Title },
				Selected: func() bool {
					c := chosen.Get()
					return c != nil && c.ID == item().ID
				},
				OnPress: func(core.PressEvent) {
					if it, ok := at(row.Index()); ok {
						chosen.Set(&it)
						s.createNav.Pop(1)
					}
				},
				StyleFunc: func() core.Style {
					c := chosen.Get()
					border := colors.Border()
					width := core.Points(1)
					if c != nil && c.ID == item().ID {
						border, width = colors.Accent(), 3
					}
					return core.Style{AspectRatio: 0.75, CornerRadius: 10, Overflow: core.Hidden, BorderWidth: width, BorderColor: border, Background: colors.Surface()}
				},
				Children: []core.Node{core.Image{ResizeMode: fit, Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)},
					SourceFunc: func() core.ImageSource { return core.Src(item().Thumb) }}},
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
				ui.Spinner{Label: title},
			}}
		}
		return core.View{Style: fill, Children: []core.Node{groups, grid}}
	})

	return ui.NavPage(cx, core.Screen{Title: title, Children: []core.Node{state}},
		ui.NavHeader{OnBack: func() { s.createNav.Pop(1) }, BackLabel: t.TabCreate()})
}

// ---- A work ----

func (s *studio) workPage(cx *core.Scope, nav *base.Nav[route], w work) base.NavPage {
	t := locales.For(cx)
	aspect := core.NewSignal(cx, float32(2.0/3.0))
	name := "pose-studio-" + w.Created.Format("20060102-150405") + filepath.Ext(w.Path)
	save := func() {
		cx.SaveFile(core.SaveOptions{Path: w.Path, Name: name}, func(saved bool, _ string, err error) {
			switch {
			case err != nil:
				s.toastError(t.SaveFailed(), err)
			case saved:
				s.toaster.Show(ui.ToastContent{Type: ui.ToastSuccess, Title: t.Saved()})
			}
		})
	}
	share := func() {
		cx.Share(core.ShareOptions{Files: []core.SharedFile{{Path: w.Path, Name: name}}}, func(bool, error) {})
	}
	remove := func() {
		base.ShowAlertDialog(cx, base.AlertDialog{Title: t.DeleteConfirm(), Description: t.DeleteConfirmBody(), Actions: []core.AlertAction{
			{Label: t.Cancel(), Role: core.AlertCancel}, {Label: t.Delete(), Role: core.AlertDestructive},
		}}, func(label string) {
			if label != t.Delete() {
				return
			}
			go func() {
				err := os.Remove(w.Path)
				if errors.Is(err, os.ErrNotExist) {
					err = nil
				}
				s.cx.Post(func() {
					if err != nil {
						s.toastError(t.Delete(), err)
						return
					}
					s.works.Update(func(ws []work) []work {
						out := make([]work, 0, len(ws))
						for _, x := range ws {
							if x.Path != w.Path {
								out = append(out, x)
							}
						}
						return out
					})
					nav.Pop(1)
				})
			}()
		})
	}
	return ui.NavPage(cx, core.Screen{Title: t.ResultTitle(), Children: []core.Node{
		core.Scroll{Style: fill, Children: []core.Node{
			core.View{Style: core.Style{Padding: core.All(16), Gap: 16}, Children: []core.Node{
				core.Image{Source: core.Src(w.Path), ResizeMode: core.Contain, TestID: "work-image", Label: t.WorkLabel(w.Created),
					OnLoad: func(px core.Size) {
						if px.Width > 0 && px.Height > 0 {
							aspect.Set(float32(px.Width / px.Height))
						}
					},
					StyleFunc: func() core.Style {
						return core.Style{Width: core.Pct(100), AspectRatio: aspect.Get(), CornerRadius: 12}
					}},
				core.View{Style: core.Style{Direction: core.Row, Gap: 10}, Children: []core.Node{
					ui.Button{Text: t.Save(), Icon: lucide.Download, Flex: 1, OnPress: save, TestID: "save"},
					ui.Button{Text: t.Share(), Icon: lucide.Share, Variant: ui.ButtonSecondary, Flex: 1, OnPress: share, TestID: "share"},
					ui.IconButton{Icon: lucide.Trash, Label: t.Delete(), Variant: ui.IconButtonGhost, OnPress: remove, TestID: "delete"},
				}},
			}},
		}},
	}}, ui.NavHeader{OnBack: func() { nav.Pop(1) }, BackLabel: t.Back()})
}

// ---- Works ----

func (s *studio) worksTab(cx *core.Scope) core.Node {
	return s.worksNav.Stack(fill, func(cx *core.Scope, r route, _ func() int) base.NavPage {
		if r.page == pageWork {
			return s.workPage(cx, s.worksNav, r.work)
		}
		return s.worksPage(cx)
	})
}

func (s *studio) worksPage(cx *core.Scope) base.NavPage {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	at := func(i int) (work, bool) {
		ws := s.works.Get()
		if i < 0 || i >= len(ws) {
			return work{}, false
		}
		return ws[i], true
	}
	return ui.NavPage(cx, core.Screen{Title: t.TabWorks(), Children: []core.Node{
		core.Show{
			When: func() bool { return len(s.works.Get()) == 0 },
			Then: func(*core.Scope) core.Node {
				return core.View{Style: fill.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Padding: core.All(24)}), Children: []core.Node{
					core.Text{Content: t.NoWorks(), TestID: "no-works", StyleFunc: func() core.Style {
						return core.Style{Color: colors.Muted(), TextAlign: core.TextAlignCenter}
					}},
				}}
			},
			Else: func(*core.Scope) core.Node {
				return core.List{Style: fill.Merge(core.Style{FlexBasis: core.Pt(0), Padding: core.All(16)}), NumColumns: 2, Gap: 10, ColumnGap: 10, EstimatedSize: 240,
					Count: func() int { return len(s.works.Get()) },
					Key: func(i int) any {
						w, _ := at(i)
						return w.Path
					},
					Row: func(cx *core.Scope, row core.ListRow) core.Node {
						current := func() work {
							w, _ := at(row.Index())
							return w
						}
						return core.View{Role: core.RoleButton,
							LabelFunc: func() string { return t.WorkLabel(current().Created) },
							OnPress: func(core.PressEvent) {
								if w, ok := at(row.Index()); ok {
									s.worksNav.Push(route{page: pageWork, work: w})
								}
							},
							StyleFunc: func() core.Style {
								return core.Style{AspectRatio: 2.0 / 3.0, CornerRadius: 12, Overflow: core.Hidden, Background: colors.Surface()}
							},
							Children: []core.Node{core.Image{ResizeMode: core.Cover, Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)},
								SourceFunc: func() core.ImageSource { return core.Src(current().Path) }}},
						}
					},
				}
			},
		},
	}}, ui.NavHeader{})
}

// ---- Settings ----

func (s *studio) settingsTab(cx *core.Scope) core.Node {
	t := locales.For(cx)
	store := cx.Storage()
	persist := func(sig *core.Signal[string], key string) func(string) {
		return func(v string) {
			sig.Set(v)
			if err := store.SetItem(key, v); err != nil {
				s.toastError(t.TabSettings(), err)
			}
		}
	}
	secrets := cx.Secrets()
	saveKey := func() {
		key := strings.TrimSpace(s.apiKey.Get())
		go func() {
			var err error
			if key == "" {
				err = secrets.Delete(apiKeyAccount)
			} else {
				err = secrets.Set(apiKeyAccount, key)
			}
			if err == nil {
				return
			}
			cx.Post(func() {
				if errors.Is(err, core.ErrSecretsUnavailable) {
					s.toaster.Show(ui.ToastContent{Type: ui.ToastWarning, Title: t.KeyUnavailable()})
					return
				}
				s.toastError(t.ApiKey(), err)
			})
		}()
	}
	field := func(title string, in core.Input) core.Node { return ui.FormField{Title: title, Input: in} }
	site := func(title, label, url string) core.Node {
		return ui.FormRow{Title: title, Detail: label, Icon: lucide.ExternalLink, OnPress: func() { cx.OpenURL(url, func(bool) {}) }}
	}
	return ui.NavPage(cx, core.Screen{Title: t.TabSettings(), Children: []core.Node{
		core.Scroll{Style: fill, Children: []core.Node{
			core.View{Style: core.Style{Padding: core.All(16), Gap: 24}, Children: []core.Node{
				ui.FormSection{Title: t.ApiSection(), Footer: t.ApiFooter(), Children: []core.Node{
					field(t.BaseURL(), core.Input{Value: s.baseURL.Get, OnChange: persist(s.baseURL, "baseURL"), Placeholder: defaultBaseURL,
						KeyboardType: core.KeyboardURL, AutoCapitalize: core.AutoCapitalizeNone, NoAutoCorrect: true, TestID: "base-url"}),
					field(t.ApiKey(), core.Input{Value: s.apiKey.Get, OnChange: s.apiKey.Set, OnBlur: saveKey, OnSubmit: func(string) { saveKey() },
						Masked: true, Placeholder: t.ApiKeyPlaceholder(), AutoCapitalize: core.AutoCapitalizeNone, NoAutoCorrect: true, TestID: "api-key"}),
					field(t.Model(), core.Input{Value: s.model.Get, OnChange: persist(s.model, "model"), Placeholder: defaultModel,
						AutoCapitalize: core.AutoCapitalizeNone, NoAutoCorrect: true, TestID: "model"}),
					field(t.Size(), core.Input{Value: s.size.Get, OnChange: persist(s.size, "size"), Placeholder: defaultSize,
						AutoCapitalize: core.AutoCapitalizeNone, NoAutoCorrect: true, TestID: "size"}),
				}},
				ui.FormSection{Title: t.SourcesSection(), Footer: t.SourcesFooter(), Children: []core.Node{
					site(t.SlotPose(), "POSE-DB-01", DefaultSites.Pose),
					site(t.SlotCostume(), "COSTUME-DB-01", DefaultSites.Costume),
					site(t.SlotFace(), "FACE REFERENCE", DefaultSites.Face),
				}},
			}},
		}},
	}}, ui.NavHeader{})
}

// loaded is the catalog res holds, empty until its first load.
func loaded(res *core.Resource[Catalog]) Catalog {
	c, _ := res.Get()
	return c
}
