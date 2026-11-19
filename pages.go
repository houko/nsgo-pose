package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/houko/nsgo-pose/locales"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func workFileName(w work) string {
	return "mirra-" + w.Created.Format("20060102-150405") + filepath.Ext(w.Path)
}

func (s *studio) saveWork(cx *core.Scope, w work) {
	t := locales.For(cx)
	cx.SaveFile(core.SaveOptions{Path: w.Path, Name: workFileName(w)}, func(saved bool, _ string, err error) {
		switch {
		case err != nil:
			s.toastError(t.SaveFailed(), err)
		case saved:
			s.toaster.Show(ui.ToastContent{Type: ui.ToastSuccess, Title: t.Saved()})
		}
	})
}

func (s *studio) shareWork(cx *core.Scope, w work) {
	cx.Share(core.ShareOptions{Files: []core.SharedFile{{Path: w.Path, Name: workFileName(w)}}}, func(bool, error) {})
}

// deleteWork asks first, then removes the picture and calls done.
func (s *studio) deleteWork(cx *core.Scope, w work, done func()) {
	t := locales.For(cx)
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
				if sh := s.shown.Get(); sh != nil && sh.Path == w.Path {
					s.shown.Set(nil)
				}
				done()
			})
		}()
	})
}

// workPage is one generated picture with save, share and delete.
func (s *studio) workPage(cx *core.Scope, nav *base.Nav[route], w work) base.NavPage {
	t := locales.For(cx)
	aspect := core.NewSignal(cx, float32(2.0/3.0))
	return ui.NavPage(cx, core.Screen{Title: t.ResultTitle(), Children: []core.Node{
		core.Scroll{Style: fill, Children: []core.Node{
			core.View{Style: core.Style{Padding: core.All(16), Gap: 16, MaxWidth: core.Pt(720), Width: core.Pct(100), AlignSelf: core.AlignCenter}, Children: []core.Node{
				core.Image{Source: core.Src(w.Path), ResizeMode: core.Contain, TestID: "work-image", Label: t.WorkLabel(w.Created),
					OnLoad: func(px core.Size) {
						if px.Width > 0 && px.Height > 0 {
							aspect.Set(float32(px.Width / px.Height))
						}
					},
					StyleFunc: func() core.Style {
						return core.Style{Width: core.Pct(100), AspectRatio: aspect.Get(), CornerRadius: 16}
					}},
				core.View{Style: core.Style{Direction: core.Row, Gap: 10}, Children: []core.Node{
					ui.Button{Text: t.Save(), Icon: lucide.Download, Flex: 1, OnPress: func() { s.saveWork(cx, w) }, TestID: "save"},
					ui.Button{Text: t.Share(), Icon: lucide.Share, Variant: ui.ButtonSecondary, Flex: 1, OnPress: func() { s.shareWork(cx, w) }, TestID: "share"},
					ui.IconButton{Icon: lucide.Trash, Label: t.Delete(), Variant: ui.IconButtonGhost, TestID: "delete",
						OnPress: func() { s.deleteWork(cx, w, func() { nav.Pop(1) }) }},
				}},
			}},
		}},
	}}, ui.NavHeader{OnBack: func() { nav.Pop(1) }, BackLabel: t.Back()})
}

func (s *studio) worksTab(cx *core.Scope) core.Node {
	return s.worksNav.Stack(fill, func(cx *core.Scope, r route, _ func() int) base.NavPage {
		if r.page == pageWork {
			return s.workPage(cx, s.worksNav, r.work)
		}
		return s.worksPage(cx)
	})
}

// worksPage is every picture generated on this device, newest first.
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
	columns := 2
	if s.wide() {
		columns = 5
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
				return core.List{Style: fill.Merge(core.Style{Padding: core.All(16)}), NumColumns: columns, Gap: 10, ColumnGap: 10, EstimatedSize: 240,
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
								return core.Style{AspectRatio: 2.0 / 3.0, CornerRadius: 14, Overflow: core.Hidden, Background: colors.Surface()}
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

// settingsPage is the image API's address, key, model and size, and the galleries the references come from.
func (s *studio) settingsPage(cx *core.Scope) core.Node {
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
