package main

import "github.com/egoist/mygo/ui"

func (a *app) toolbar(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	bar := c.TitleBar()
	var modelButton ui.Element
	ui.Row(c).FillWidth().Height(48).Padding(0, bar.Right+10, 0, bar.Left+8).
		Gap(8).AlignItems(ui.Center).Background(p.surface2).
		BorderWidth(0, 0, 1, 0).BorderColor(p.hairline).Children(func() {
		ui.Text(c, "✦").FontSize(12).TextColor(p.blue)
		ui.Text(c, "JustLearn").FontSize(13).FontWeight(600).TextColor(p.text)
		ui.Box(c).Width(1).Height(16).Background(p.hairline)
		title := a.goal.Title
		if title == "" {
			title = a.active.Title
		}
		if title == "" {
			title = "新对话"
		}
		ui.Text(c, title).FontSize(fsUI).FontWeight(500).TextColor(p.muted).
			MaxLines(1).MaxWidth(320).DragWindow()
		ui.Box(c).Grow(1).FillHeight().DragWindow()

		modelButton = ui.ButtonBase(c).Padding(3, 8).Radius(6).
			Label("选择模型").Tooltip("选择模型").Children(func() {
			ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
				ui.Text(c, prettyModel(cfgModel(a.store))).FontSize(fsUI).TextColor(p.muted).MaxLines(1).MaxWidth(180)
				ui.Text(c, "⌄").FontSize(10).TextColor(p.faint)
			})
		})
		if modelButton.Hovered() {
			modelButton.Background(p.navHover)
		}
		if modelButton.Clicked() {
			a.modelMenuOpen = !a.modelMenuOpen
			if a.modelMenuOpen {
				a.loadModels()
			}
		}

		settings := ui.ButtonBase(c).Padding(3, 8).Radius(6).Label("设置").Tooltip("设置").Children(func() {
			ui.Text(c, "设置").FontSize(fsUI).TextColor(p.muted)
		})
		if settings.Hovered() {
			settings.Background(p.navHover)
		}
		if settings.Clicked() {
			a.openSettings()
		}
	})
	ui.PopoverBase(c, modelButton, &a.modelMenuOpen, func(panel ui.Element) {
		a.modelPanel(c, t, p, panel)
	})
}
