package main

import "github.com/egoist/mygo/ui"

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).Width(208).FillHeight().Background(p.sidebar).
		BorderWidth(0, 1, 0, 0).BorderColor(p.navLine).Children(func() {
		// Brand sits in the window's title band, next to the traffic lights.
		bar := c.TitleBar()
		ui.Row(c).Height(max(bar.Height, 40)).Padding(0, bar.Right+8, 0, bar.Left+8).
			Gap(6).AlignItems(ui.Center).DragWindow().Children(func() {
			ui.Box(c).Size(18, 18).Radius(4).Background(p.navActive).Center().Children(func() {
				ui.Text(c, "✦").FontSize(10).TextColor(p.blue)
			})
			ui.Text(c, "JustLearn").FontSize(13).FontWeight(600).TextColor(p.sidebarText)
		})

		// Search routes
		ui.Column(c).Padding(2, 6, 2, 6).Children(func() {
			ui.SearchField(c, &a.search).FillWidth().Placeholder("搜索路线")
		})

		// New route
		ui.Column(c).PaddingX(6).Children(func() {
			btn := ui.ButtonBase(c).FillWidth().Height(24).Radius(radiusBadge).
				PaddingX(6).Justify(ui.Start).Children(func() {
				ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "+").FontSize(13).TextColor(p.muted)
					ui.Text(c, "新路线").FontSize(fsUI).TextColor(p.muted)
				})
			})
			if btn.Hovered() {
				btn.Background(p.navHover)
			}
			if btn.Clicked() {
				a.newRoute()
			}
		})

		// Routes
		ui.Text(c, "学习路线").FontSize(fsCaption).TextColor(p.faint).
			Padding(10, 10, 2, 10)
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().PaddingX(6).PaddingY(2).Gap(1).Children(func() {
				routes := a.visibleRoutes()
				if len(routes) == 0 {
					ui.Text(c, "还没有路线").FontSize(fsCaption).TextColor(p.faint).Padding(6, 4)
				}
				for _, goal := range routes {
					g := goal
					active := g.ID == a.goal.ID
					weight := 400
					if active {
						weight = 600
					}
					wrap := ui.Row(c).FillWidth().Gap(2).AlignItems(ui.Center)
					wrap.Children(func() {
						row := ui.ButtonBase(c.Key("route:" + g.ID)).Grow(1).Height(26).Radius(radiusBadge).
							PaddingX(6).Children(func() {
							ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
								ui.Text(c, g.Title).FontSize(fsUI).FontWeight(weight).MaxLines(1).
									TextColor(p.sidebarText)
								ui.Spacer(c)
								ui.Textf(c, "%d%%", g.Percent()).FontSize(fsCaption).TextColor(p.faint)
							})
						})
						if active {
							row.Background(p.navActive)
						} else if row.Hovered() {
							row.Background(p.navHover)
						}
						if row.Clicked() {
							a.selectGoal(g.ID)
							a.routeNodeID = ""
						}
						del := ui.ButtonBase(c).Size(18, 18).Radius(radiusBadge).Label("删除路线").
							Tooltip("删除路线").Center().Children(func() {
							ui.Text(c, "×").FontSize(11).TextColor(p.faint)
						})
						if del.Hovered() {
							del.Background(p.fieldBg)
						}
						if del.Clicked() {
							a.askDeleteRoute(g.ID)
						}
					})
				}
			})
		})

		// Footer
		ui.Row(c).Padding(4, 8, 6, 8).Gap(8).AlignItems(ui.Center).Children(func() {
			settings := ui.ButtonBase(c).Padding(3, 4).Radius(radiusBadge).Children(func() {
				ui.Text(c, "设置").FontSize(fsCaption).TextColor(p.faint)
			})
			if settings.Hovered() {
				settings.Background(p.navHover)
			}
			if settings.Clicked() {
				a.openSettings()
			}
		})
	})
}
