package main

import "github.com/egoist/mygo/ui"

// toolbar is the tiny cluster in the top-right corner. The ability profile
// is already fixed in the right margin, so settings is the only control.
func (a *app) toolbar(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Row(c).Absolute().Top(8).Right(8).Gap(4).AlignItems(ui.Center).Children(func() {
		settings := iconButton(c, p, iconConfig, "设置")
		if settings.Clicked() {
			a.toggleSettingsPanel()
		}
	})
}
