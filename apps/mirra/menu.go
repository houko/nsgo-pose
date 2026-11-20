package main

import (
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
	"github.com/nsgo-dev/nsgo/updater"
	"github.com/yldm-ai/mirra/apps/mirra/locales"
)

// menu is the macOS menu bar: the app menu with About, Settings… and, where the app updates itself, Check for Updates…, then Edit and Window. NSGo's default menu has About and Check for Updates… but no Settings…, an app's own, and none of AppKit's keys for Hide and Minimize (https://github.com/nsgo-dev/nsgo/issues/947), so the app sets the whole bar with the roles, which carry those keys.
func (s *studio) menu(cx *core.Scope, t locales.Messages) []desktop.MenuItem {
	app := []desktop.MenuItem{
		{Role: desktop.MenuAbout}, {Separator: true},
		{Label: t.MenuSettings(), Shortcut: "mod+,", OnSelect: s.openSettings},
	}
	if updatable(cx.Env()) {
		app = append(app, desktop.MenuItem{Label: t.MenuCheckForUpdates(), OnSelect: func() { updater.Check(cx) }})
	}
	app = append(app, desktop.MenuItem{Separator: true},
		desktop.MenuItem{Role: desktop.MenuServices}, desktop.MenuItem{Separator: true},
		desktop.MenuItem{Role: desktop.MenuHide}, desktop.MenuItem{Role: desktop.MenuHideOthers}, desktop.MenuItem{Role: desktop.MenuShowAll}, desktop.MenuItem{Separator: true},
		desktop.MenuItem{Role: desktop.MenuQuit})
	return []desktop.MenuItem{
		{Label: t.AppName(), Items: app},
		{Label: t.MenuEdit(), Items: []desktop.MenuItem{
			{Role: desktop.MenuUndo}, {Role: desktop.MenuRedo}, {Separator: true},
			{Role: desktop.MenuCut}, {Role: desktop.MenuCopy}, {Role: desktop.MenuPaste}, {Role: desktop.MenuSelectAll},
		}},
		{Label: t.MenuWindow(), Role: desktop.MenuWindowMenu, Items: []desktop.MenuItem{
			{Role: desktop.MenuMinimize}, {Role: desktop.MenuZoom}, {Separator: true}, {Role: desktop.MenuBringAllToFront},
		}},
	}
}
