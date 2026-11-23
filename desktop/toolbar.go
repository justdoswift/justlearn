package main

import "github.com/egoist/mygo/ui"

// toolbar is the tiny cluster in the top-right corner. The ability profile is
// already fixed in the right margin, so only navigation and settings live
// here.
func (a *app) toolbar(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Row(c).Absolute().Top(8).Right(8).Gap(2).AlignItems(ui.Center).Children(func() {
		var navButton ui.Element
		navButton = iconButton(c, p, iconSidebar, "路线与对话")
		if navButton.Clicked() {
			a.navOpen = !a.navOpen
			if a.navOpen {
				a.settingsOpen = false
				a.modelMenuOpen = false
			}
		}
		ui.PopoverBase(c, navButton, &a.navOpen, func(panel ui.Element) {
			a.navPanel(c, t, p, panel, navButton)
		})

		settings := iconButton(c, p, iconConfig, "设置")
		if settings.Clicked() {
			a.openSettings()
		}
	})
}
