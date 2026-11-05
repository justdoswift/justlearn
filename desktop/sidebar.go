package main

import "github.com/egoist/mygo/ui"

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).Width(208).FillHeight().Background(p.sidebar).Children(func() {
		// Brand sits in the window's title band, next to the traffic lights.
		bar := c.TitleBar()
		ui.Row(c).Height(max(bar.Height, 40)).Padding(0, bar.Right+8, 0, bar.Left+8).
			Gap(6).AlignItems(ui.Center).DragWindow().Children(func() {
			ui.Box(c).Size(18, 18).Radius(4).Background(p.navActive).Center().Children(func() {
				ui.Text(c, "✦").FontSize(10).TextColor(p.sidebarText)
			})
			ui.Text(c, "JustLearn").FontSize(13).FontWeight(600).TextColor(p.sidebarText)
		})

		// New conversation
		ui.Column(c).PaddingX(6).Children(func() {
			btn := ui.ButtonBase(c).FillWidth().Height(24).Radius(radiusBadge).
				PaddingX(6).Justify(ui.Start).Children(func() {
				ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
					ui.Text(c, "+").FontSize(13).TextColor(p.sidebarMuted)
					ui.Text(c, "新对话").FontSize(fsUI).TextColor(p.sidebarMuted)
				})
			})
			if btn.Hovered() {
				btn.Background(p.navHover)
			}
			if btn.Clicked() {
				a.newConversation()
			}
		})

		// Goal, as one quiet line with a progress rule.
		if a.goal.ID != "" {
			title := a.goal.Title
			if title == "" {
				title = "新的学习路线"
			}
			ui.Column(c).Padding(6, 10, 4, 10).Gap(3).Children(func() {
				ui.Row(c).FillWidth().Children(func() {
					ui.Text(c, title).FontSize(fsCaption).TextColor(p.sidebarMuted).MaxLines(1)
					ui.Spacer(c)
					ui.Textf(c, "%d%%", a.goal.Percent()).FontSize(fsCaption).TextColor(p.sidebarText)
				})
				ui.Row(c).FillWidth().Height(2).Radius(1).Background(p.navLine).Children(func() {
					ui.Box(c).Height(2).Radius(1).Background(p.blue).
						WidthPercent(float32(a.goal.Percent()))
				})
			})
		}

		// Conversations, one line each.
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().PaddingX(6).PaddingY(4).Children(func() {
				for _, group := range groupConversations(a.convs) {
					ui.Text(c, group.label).FontSize(fsCaption).TextColor(p.faint).
						Padding(6, 4, 2, 6)
					for _, conv := range group.items {
						active := conv.ID == a.active.ID
						weight := 400
						if active {
							weight = 600
						}
						row := ui.ButtonBase(c.Key(conv.ID)).FillWidth().Height(24).Radius(radiusBadge).
							PaddingX(6).Justify(ui.Start).Children(func() {
							ui.Text(c, conv.Title).FontSize(fsUI).FontWeight(weight).
								TextColor(p.sidebarText).MaxLines(1)
						})
						if active {
							row.Background(p.navActive)
						} else if row.Hovered() {
							row.Background(p.navHover)
						}
						if row.Clicked() && !a.streaming {
							a.selectConversation(conv.ID)
						}
					}
				}
			})
		})

		// Footer
		ui.Row(c).Padding(4, 8, 6, 8).Gap(8).AlignItems(ui.Center).Children(func() {
			route := ui.ButtonBase(c).Padding(3, 4).Radius(radiusBadge).Children(func() {
				color := p.faint
				if a.routeOpen {
					color = p.sidebarText
				}
				ui.Text(c, "路线").FontSize(fsCaption).TextColor(color)
			})
			if route.Hovered() {
				route.Background(p.navHover)
			}
			if route.Clicked() {
				a.routeOpen = !a.routeOpen
			}
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
