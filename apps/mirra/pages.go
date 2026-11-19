package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yldm-ai/mirra/apps/mirra/locales"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// ---- A shot ----

// shotEyebrow says what the picture on show is: a new try, by number, or a work in the lookbook.
func shotEyebrow(t locales.Messages, sh *shot) string {
	if sh.Saved != nil {
		return t.EyebrowLookbookNo(workNo(sh.Saved.No))
	}
	return t.EyebrowNewLook(sh.Attempt)
}

// swapOutfits are the outfits one tap puts on the same face and pose: the favorites, the user's clothes, then the gallery's in the chip on show, without the one being worn.
func (s *studio) swapOutfits() []Item {
	var out []Item
	seen := map[string]bool{}
	if sh := s.shot.Get(); sh != nil && sh.Req.Costume != nil {
		seen[sh.Req.Costume.ID] = true
	}
	add := func(items []Item) {
		for _, it := range items {
			if len(out) == 8 {
				return
			}
			if !seen[it.ID] {
				seen[it.ID] = true
				out = append(out, it)
			}
		}
	}
	add(s.favorites.Get())
	add(s.uploads.Get())
	add(loaded(s.catalogs[KindCostume]).InGroup(s.filter[KindCostume].Get()))
	return out
}

func (s *studio) swapRow(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	return core.View{Style: core.Style{Gap: 10}, Children: []core.Node{
		caption(cx, t.SwapOutfit),
		core.Scroll{Horizontal: true, HideIndicator: true, Style: core.Style{Height: core.Pt(84), MinHeight: core.Pt(84), FlexShrink: 0}, Children: []core.Node{
			core.Dynamic(func(cx *core.Scope) core.Node {
				var nodes []core.Node
				for i, it := range s.swapOutfits() {
					it := it
					nodes = append(nodes, core.View{Role: core.RoleButton, Label: t.SwapTo(itemName(t, &it)), TestID: "swap-" + string(rune('0'+i)),
						OnPress: func(core.PressEvent) { s.regenerate(&it) },
						StyleFunc: func() core.Style {
							return core.Style{Width: core.Pt(60), Height: core.Pt(80), CornerRadius: 12, Overflow: core.Hidden, Background: colors.Inset(),
								BorderWidth: 1, BorderColor: colors.Border()}
						},
						Children: []core.Node{core.Image{Source: core.Src(it.Thumb), ResizeMode: core.Contain, Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}}},
					})
				}
				return core.View{Style: core.Style{Direction: core.Row, Gap: 8}, Children: nodes}
			}),
		}},
	}}
}

// usedRefs are the face and pose a shot was made from, small, over the picture's corner.
func (s *studio) usedRefs(cx *core.Scope, sh *shot) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	var nodes []core.Node
	if sh.Req.Face != nil {
		nodes = append(nodes, core.View{Label: t.UsedFace(), Role: core.RoleImage, Children: []core.Node{thumb(cx, KindFace, func() *Item { return sh.Req.Face }, 48, 48, true)}})
	}
	if sh.Req.Pose != nil {
		nodes = append(nodes, core.View{Label: t.UsedPose(), Role: core.RoleImage, Children: []core.Node{thumb(cx, KindPose, func() *Item { return sh.Req.Pose }, 48, 64, false)}})
	}
	return core.View{StyleFunc: func() core.Style {
		return core.Style{Direction: core.Row, Gap: 8, AlignItems: core.AlignEnd, Padding: core.All(4), CornerRadius: 16, Background: colors.Background().Faded(0.6)}
	}, Children: nodes}
}

// saveButton keeps the shot in Works, or says it is there.
func (s *studio) saveButton(cx *core.Scope, sh *shot) core.Node {
	t := locales.For(cx)
	if sh.Saved != nil {
		return ui.Button{Text: t.InWorks(), Icon: lucide.Check, Variant: ui.ButtonSecondary, Flex: 1, TestID: "saved", Disabled: func() bool { return true },
			Style: core.Style{CornerRadius: 999}}
	}
	return ui.Button{Text: t.SaveToWorks(), Flex: 1, TestID: "save-to-works", Style: core.Style{CornerRadius: 999}, OnPress: func() { s.saveShot(false) }}
}

func (s *studio) favoriteButton(cx *core.Scope, sh *shot) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	fav := sh.Saved != nil && sh.Saved.Favorite
	label := t.Favorite()
	if fav {
		label = t.Unfavorite()
	}
	return ui.IconButton{Icon: lucide.Heart, Label: label, Variant: ui.IconButtonRaised, Selected: func() bool { return fav }, TestID: "favorite",
		Tint: func() core.Color {
			if fav {
				return colors.Danger()
			}
			return colors.Text()
		}, OnPress: s.toggleFavorite}
}

// shotPage is the phone's result: the picture, its actions over it, then its title, the outfits to swap to, Generate again and Save.
func (s *studio) shotPage(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	return core.Dynamic(func(cx *core.Scope) core.Node {
		sh := s.shot.Get()
		if sh == nil {
			return nil
		}
		picture := core.View{Style: fill, Children: []core.Node{
			core.Image{Source: core.Src(sh.Path), ResizeMode: core.Contain, TestID: "shot-image", Label: lookTitle(t, sh.Req), Style: core.Style{Width: core.Pct(100), Height: core.Pct(100)}},
			core.View{Style: core.Style{Position: core.Absolute, Top: core.Pt(12), Left: core.Pt(16), Right: core.Pt(16), Direction: core.Row, Gap: 10}, Children: []core.Node{
				ui.IconButton{Icon: lucide.X, Label: t.Close(), Variant: ui.IconButtonRaised, TestID: "close-shot", OnPress: s.closeShot},
				core.View{Style: core.Style{FlexGrow: 1}},
				ui.IconButton{Icon: lucide.Download, Label: t.Export(), Variant: ui.IconButtonRaised, TestID: "export", OnPress: func() { s.saveFile(cx, sh.Path) }},
				ui.IconButton{Icon: lucide.Trash, Label: t.Delete(), Variant: ui.IconButtonRaised, TestID: "delete", OnPress: s.discardShot},
				s.favoriteButton(cx, sh),
			}},
			core.View{Style: core.Style{Position: core.Absolute, Right: core.Pt(16), Bottom: core.Pt(12)}, Children: []core.Node{s.usedRefs(cx, sh)}},
		}}
		return core.View{TestID: "shot", StyleFunc: func() core.Style { return fill.Merge(core.Style{Background: colors.Background()}) }, Children: []core.Node{
			picture,
			core.View{Style: core.Style{Padding: core.Symmetric(14, 20), Gap: 14, PaddingEnv: core.EnvEdges{Bottom: core.EnvSafeArea}}, Children: []core.Node{
				core.View{Style: core.Style{Gap: 4}, Children: []core.Node{
					eyebrow(cx, func() string { return shotEyebrow(t, sh) }),
					display(func() string { return lookTitle(t, sh.Req) }, 26),
				}},
				s.swapRow(cx),
				core.View{Style: core.Style{Direction: core.Row, Gap: 10}, Children: []core.Node{
					ui.Button{Text: t.Regenerate(), Icon: lucide.RefreshCw, Variant: ui.ButtonOutline, Flex: 1, TestID: "regenerate", Style: core.Style{CornerRadius: 999},
						OnPress: func() { s.regenerate(nil) }},
					s.saveButton(cx, sh),
				}},
			}},
		}}
	})
}

// ---- Works ----

// worksPage is the lookbook: every saved picture, newest first, or only the favorites.
func (s *studio) worksPage(cx *core.Scope, columns int) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	shown := core.NewMemoFunc(cx, func() []work {
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
	}, nil)
	at := func(i int) (work, bool) {
		ws := shown.Get()
		if i < 0 || i >= len(ws) {
			return work{}, false
		}
		return ws[i], true
	}
	favCount := func() int {
		n := 0
		for _, w := range s.works.Get() {
			if w.Favorite {
				n++
			}
		}
		return n
	}
	filter := ui.Tabs[bool]{Label: t.WorksFilter(), Value: s.favoritesOn.Get, OnChange: s.favoritesOn.Set, TestID: "works-filter", Options: []ui.Option[bool]{
		{Value: false, Label: t.All(), TestID: "works-all", Content: core.Text{ContentFunc: func() string { return t.AllCount(len(s.works.Get())) }}},
		{Value: true, Label: t.Favorites(), TestID: "works-favorites", Content: core.Text{ContentFunc: func() string { return t.FavoritesCount(favCount()) }}},
	}}
	empty := func(cx *core.Scope) core.Node {
		title, body := t.NoWorksTitle(), t.NoWorksBody()
		if s.favoritesOn.Get() {
			title, body = t.NoFavoriteWorksTitle(), ""
		}
		children := []core.Node{core.Text{Content: title, Style: core.Style{FontFamily: displayFont, FontSize: 22, FontWeight: core.FontWeightBold}}}
		if body != "" {
			children = append(children, core.Text{Content: body, StyleFunc: func() core.Style {
				return core.Style{FontSize: 14, Color: colors.Muted(), TextAlign: core.TextAlignCenter}
			}}, ui.Button{Text: t.FirstLook(), TestID: "first-look", Style: core.Style{CornerRadius: 999, Margin: core.Edges{Top: 8}},
				OnPress: func() { s.section.Set(sectionStudio) }})
		}
		return core.View{TestID: "no-works", Style: fill.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Gap: 8, Padding: core.All(24)}), Children: children}
	}
	grid := core.List{Style: fill, NumColumns: columns, Gap: 12, ColumnGap: 12, EstimatedSize: 300,
		Count: func() int { return len(shown.Get()) },
		Key: func(i int) any {
			w, _ := at(i)
			return w.Path
		},
		Row: func(cx *core.Scope, row core.ListRow) core.Node {
			current := func() work {
				w, _ := at(row.Index())
				return w
			}
			return core.View{Role: core.RoleButton, LabelFunc: func() string { return t.OpenWork(workNo(current().No)) },
				OnPress: func(core.PressEvent) {
					if w, ok := at(row.Index()); ok {
						s.openWork(w)
					}
				},
				StyleFunc: func() core.Style {
					return core.Style{CornerRadius: 16, Overflow: core.Hidden, Background: colors.Surface(), BorderWidth: 1, BorderColor: colors.Border()}
				},
				Children: []core.Node{
					core.Image{ResizeMode: core.Cover, Style: core.Style{Width: core.Pct(100), AspectRatio: 2.0 / 3.0},
						SourceFunc: func() core.ImageSource { return core.Src(current().Path) }},
					core.View{Style: core.Style{Padding: core.Symmetric(8, 10), Gap: 2}, Children: []core.Node{
						core.Text{ContentFunc: func() string { return t.EyebrowLookbookNo(workNo(current().No)) }, StyleFunc: func() core.Style {
							return core.Style{FontSize: 10, LetterSpacing: 1.5, Color: colors.Accent()}
						}},
						core.Text{ContentFunc: func() string {
							w := current()
							name := lookTitle(t, Request{Costume: w.Costume, Pose: w.Pose})
							if w.Favorite {
								name = "♥ " + name
							}
							return name
						}, MaxLines: 1, Style: core.Style{FontSize: 14, FontWeight: core.FontWeightMedium}},
					}},
				},
			}
		},
	}
	return core.View{TestID: "works", Style: fill.Merge(core.Style{Gap: 14}), Children: []core.Node{
		core.View{Style: core.Style{Gap: 4}, Children: []core.Node{eyebrow(cx, t.EyebrowLookbook), display(t.TabWorks, 30)}},
		filter,
		core.Show{When: func() bool { return len(shown.Get()) == 0 }, Then: empty, Else: func(*core.Scope) core.Node { return grid }},
	}}
}

// ---- AI service ----

func (s *studio) resetConnection() {
	s.conn.Set(connUntested)
	s.connErr.Set("")
}

func (s *studio) testConnection() {
	cfg, client := s.config(), s.Client
	s.conn.Set(connTesting)
	s.connErr.Set("")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := CheckConnection(ctx, client, cfg)
		s.cx.Post(func() {
			if err != nil {
				s.conn.Set(connFailed)
				s.connErr.Set(err.Error())
				return
			}
			s.conn.Set(connOK)
		})
	}()
}

// field is a labeled input as a column: the label above, then the field.
func field(cx *core.Scope, label func() string, in core.Node) core.Node {
	return core.View{Style: core.Style{Gap: 8}, Children: []core.Node{caption(cx, label), in}}
}

// serviceSettings sets up the AI service: cloud or local, its address and key, model and size, and a test of the connection.
func (s *studio) serviceSettings(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	store := cx.Storage()
	secrets := cx.Secrets()
	local := func() bool { return s.service.Get() == serviceLocal }
	url := func() string {
		if local() {
			return s.localURL.Get()
		}
		return s.cloudURL.Get()
	}
	setURL := func(v string) {
		sig, key := s.cloudURL, keyCloudURL
		if local() {
			sig, key = s.localURL, keyLocalURL
		}
		sig.Set(v)
		s.resetConnection()
		if err := store.SetItem(key, v); err != nil {
			s.toastError(t.SaveFailed(), err)
		}
	}
	apiKey := func() string {
		if local() {
			return s.localKey.Get()
		}
		return s.cloudKey.Get()
	}
	setKey := func(v string) {
		if local() {
			s.localKey.Set(v)
		} else {
			s.cloudKey.Set(v)
		}
		s.resetConnection()
	}
	saveKey := func() {
		account, key := apiKeyAccount, strings.TrimSpace(s.cloudKey.Get())
		if local() {
			account, key = localKeyAccount, strings.TrimSpace(s.localKey.Get())
		}
		go func() {
			var err error
			if key == "" {
				err = secrets.Delete(account)
			} else {
				err = secrets.Set(account, key)
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
	setStored := func(sig *core.Signal[string], key string) func(string) {
		return func(v string) {
			sig.Set(v)
			s.resetConnection()
			if err := store.SetItem(key, v); err != nil {
				s.toastError(t.SaveFailed(), err)
			}
		}
	}
	keyLabel := func() string {
		if local() {
			return t.ApiKeyOptional()
		}
		return t.ApiKey()
	}
	keyField := core.View{Style: core.Style{Direction: core.Row, Gap: 8}, Children: []core.Node{
		core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1}, Children: []core.Node{
			core.Dynamic(func(cx *core.Scope) core.Node {
				return ui.Input{Value: apiKey, OnChange: setKey, OnBlur: saveKey, Masked: !s.showKey.Get(), Label: keyLabel(), Placeholder: t.ApiKeyPlaceholder(), TestID: "api-key"}
			}),
		}},
		core.Dynamic(func(cx *core.Scope) core.Node {
			text := t.ShowKey()
			if s.showKey.Get() {
				text = t.HideKey()
			}
			return ui.Button{Variant: ui.ButtonOutline, TestID: "show-key", Style: core.Style{CornerRadius: 12},
				Text: text, OnPress: func() { s.showKey.Update(func(v bool) bool { return !v }) }}
		}),
	}}
	result := core.Dynamic(func(cx *core.Scope) core.Node {
		switch s.conn.Get() {
		case connOK:
			return core.Text{Content: "✓ " + t.ConnectionOK(), TestID: "conn-ok", StyleFunc: func() core.Style { return core.Style{FontSize: 14, Color: colors.Up()} }}
		case connFailed:
			return core.Text{Content: t.ConnectionFailed(s.connErr.Get()), TestID: "conn-failed", StyleFunc: func() core.Style { return core.Style{FontSize: 14, Color: colors.Danger()} }}
		}
		return nil
	})
	return core.View{Style: core.Style{Gap: 16}, Children: []core.Node{
		core.Show{When: func() bool { return !s.configured() }, Then: func(cx *core.Scope) core.Node {
			return core.View{StyleFunc: func() core.Style {
				return core.Style{Padding: core.Symmetric(10, 14), CornerRadius: 12, BorderWidth: 1, BorderColor: colors.Accent().Faded(0.4), Background: colors.Accent().Faded(0.08)}
			}, Children: []core.Node{core.Text{Content: t.ServiceIntro(), StyleFunc: func() core.Style { return core.Style{FontSize: 14, Color: colors.Accent()} }}}}
		}},
		ui.Tabs[serviceKind]{Label: t.ServiceKind(), Equal: true, Value: s.service.Get, TestID: "service-kind",
			OnChange: func(k serviceKind) {
				s.service.Set(k)
				s.resetConnection()
				if err := store.SetItem(keyService, string(k)); err != nil {
					s.toastError(t.SaveFailed(), err)
				}
			},
			Options: []ui.Option[serviceKind]{{Value: serviceCloud, Label: t.ServiceCloud(), TestID: "service-cloud"}, {Value: serviceLocal, Label: t.ServiceLocal(), TestID: "service-local"}}},
		field(cx, t.BaseURL, core.Dynamic(func(cx *core.Scope) core.Node {
			placeholder := defaultBaseURL
			if local() {
				placeholder = defaultLocalURL
			}
			return ui.Input{Value: url, OnChange: setURL, Label: t.BaseURL(), Placeholder: placeholder, KeyboardType: core.KeyboardURL, TestID: "base-url"}
		})),
		field(cx, keyLabel, keyField),
		core.View{Style: core.Style{Direction: core.Row, Gap: 10}, Children: []core.Node{
			core.View{Style: core.Style{FlexGrow: 1, FlexBasis: core.Pt(0)}, Children: []core.Node{
				field(cx, t.Model, ui.Input{Value: s.model.Get, OnChange: setStored(s.model, keyModel), Label: t.Model(), Placeholder: defaultModel, TestID: "model"}),
			}},
			core.View{Style: core.Style{FlexGrow: 1, FlexBasis: core.Pt(0)}, Children: []core.Node{
				field(cx, t.Size, ui.Input{Value: s.size.Get, OnChange: setStored(s.size, keySize), Label: t.Size(), Placeholder: defaultSize, TestID: "size"}),
			}},
		}},
		core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 12}, Children: []core.Node{
			core.Dynamic(func(cx *core.Scope) core.Node {
				text := t.TestConnection()
				if s.conn.Get() == connOK || s.conn.Get() == connFailed {
					text = t.TestAgain()
				}
				return ui.Button{Text: text, TestID: "test-connection", Style: core.Style{CornerRadius: 999},
					Busy: func() bool { return s.conn.Get() == connTesting }, Disabled: func() bool { return strings.TrimSpace(url()) == "" }, OnPress: s.testConnection}
			}),
		}},
		result,
		core.Text{ContentFunc: func() string {
			if local() {
				return t.PrivacyLocal()
			}
			return t.PrivacyCloud()
		}, StyleFunc: func() core.Style { return core.Style{FontSize: 13, Color: colors.Muted()} }},
	}}
}

// aboutLinks are the app's source, its license and where its references come from.
func (s *studio) aboutLinks(cx *core.Scope) core.Node {
	t := locales.For(cx)
	open := func(url string) func() { return func() { cx.OpenURL(url, func(bool) {}) } }
	return ui.FormSection{Title: t.AboutSection(), Children: []core.Node{
		ui.FormRow{Title: t.SourceCode(), Detail: "GitHub", OnPress: open(repoURL)},
		ui.FormRow{Title: t.License(), Detail: "MIT", OnPress: open(repoURL + "/blob/main/LICENSE")},
		ui.FormRow{Title: t.SourcesPose(), Detail: "POSE-DB-01", OnPress: open(DefaultSites.Pose)},
		ui.FormRow{Title: t.SourcesOutfit(), Detail: "COSTUME-DB-01", OnPress: open(DefaultSites.Costume)},
		ui.FormRow{Title: t.SourcesFace(), Detail: "FACE REFERENCE · CC0", OnPress: open(DefaultSites.Face)},
	}, Footer: t.SourcesFooter()}
}

// card is a rounded panel on the page.
func card(cx *core.Scope, children ...core.Node) core.Node {
	colors := ui.ColorsOf(cx)
	return core.View{StyleFunc: func() core.Style {
		return core.Style{Padding: core.All(20), Gap: 14, CornerRadius: 24, Background: colors.Surface(), BorderWidth: 1, BorderColor: colors.Border()}
	}, Children: children}
}

// appCard is the app's name and terms with the service's status.
func (s *studio) appCard(cx *core.Scope) core.Node {
	t := locales.For(cx)
	colors := ui.ColorsOf(cx)
	return card(cx, core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 14}, Children: []core.Node{
		core.View{StyleFunc: func() core.Style {
			return core.Style{Width: core.Pt(56), Height: core.Pt(56), CornerRadius: 16, Background: colors.Text(), AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter}
		}, Children: []core.Node{core.Text{Content: "M", StyleFunc: func() core.Style {
			return core.Style{FontFamily: displayFont, FontSize: 26, FontWeight: core.FontWeightBold, Color: colors.Background()}
		}}}},
		core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, Gap: 4}, Children: []core.Node{
			core.Text{Content: t.Brand() + " " + t.BrandTagline(), Style: core.Style{FontSize: 18, FontWeight: core.FontWeightBold}},
			caption(cx, t.OpenSourceTerms),
		}},
	}}, s.statusBadge(cx))
}

// settingsPage is the phone's Me tab: the app, its AI service and where it comes from.
func (s *studio) settingsPage(cx *core.Scope) core.Node {
	t := locales.For(cx)
	return core.Screen{HeaderHidden: true, Title: t.TabMe(), Style: fill, Children: []core.Node{
		core.Scroll{Style: fill, Children: []core.Node{
			core.View{TestID: "settings", Style: core.Style{Padding: core.All(20), Gap: 16, PaddingEnv: core.EnvEdges{Top: core.EnvSafeArea}}, Children: []core.Node{
				core.View{Style: core.Style{Gap: 4}, Children: []core.Node{eyebrow(cx, t.EyebrowSettings), display(t.TabMe, 30)}},
				s.appCard(cx),
				card(cx, display(t.AiService, 20), s.serviceSettings(cx)),
				s.aboutLinks(cx),
			}},
		}},
	}}
}

// settingsDialog is the desktop's AI service settings.
func (s *studio) settingsDialog(cx *core.Scope) core.Node {
	t := locales.For(cx)
	return ui.Dialog{Open: s.settingsShown.Get, OnOpenChange: func(open bool, _ base.Reason) { s.settingsShown.Set(open) },
		Title: t.AiServiceSettings(), TestID: "settings-dialog", Style: core.Style{Width: core.Pt(560)},
		Footer: []core.Node{ui.Button{Text: t.Close(), Variant: ui.ButtonSecondary, TestID: "close-settings", Style: core.Style{CornerRadius: 999},
			OnPress: func() { s.settingsShown.Set(false) }}},
		Children: []core.Node{
			core.View{TestID: "settings", Style: core.Style{Gap: 16}, Children: []core.Node{s.serviceSettings(cx), s.aboutLinks(cx)}},
		}}
}
