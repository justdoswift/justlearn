package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *app) sidebar(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).Width(280).FillHeight().Background(p.sidebar).
		BorderWidth(0, 1, 0, 0).BorderColor(p.navLine).Children(func() {
		ui.Column(c).Padding(6, 8, 4, 8).Children(func() {
			ui.SearchField(c, &a.search).FillWidth().Height(28).Radius(6).
				Background(p.fieldBg).BorderWidth(1).BorderColor(p.hairline).Placeholder("搜索节点")
		})

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

		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().Padding(6, 6, 4, 6).Gap(1).Children(func() {
				ui.Text(c, "路线").FontSize(fsCaption).TextColor(p.faint).Padding(4, 4, 2, 4)
				routes := a.goals
				if len(routes) == 0 {
					ui.Text(c, "还没有路线").FontSize(fsCaption).TextColor(p.faint).Padding(4, 4)
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
								pct := p.faint
								if active {
									pct = p.muted
								}
								ui.Text(c, g.Title).FontSize(fsRead).FontWeight(weight).MaxLines(1).
									TextColor(p.sidebarText)
								ui.Spacer(c)
								ui.Textf(c, "%d%%", g.Percent()).FontSize(fsCaption).TextColor(pct)
							})
						})
						if active {
							row.Background(p.navHover)
							row.BorderWidth(0, 0, 0, 2).BorderColor(p.blue)
						} else if row.Hovered() {
							row.Background(p.navHover)
						}

						if row.Clicked() {
							a.selectGoal(g.ID)
							a.routeNodeID = ""
						}
						xColor := p.faint
						if active {
							xColor = p.muted
						}
						del := ui.ButtonBase(c).Size(18, 18).Radius(radiusBadge).Label("删除路线").
							Tooltip("删除路线").Center().Children(func() {
							ui.Text(c, "×").FontSize(11).TextColor(xColor)
						})
						if del.Hovered() {
							del.Background(p.fieldBg)
						}
						if del.Clicked() {
							a.askDeleteRoute(g.ID)
						}
					})
				}

				hairline(c, p.hairline)
				ui.Text(c, "节点").FontSize(fsCaption).TextColor(p.faint).Padding(6, 4, 2, 4)

				nodes := a.visibleNodes()
				if len(nodes) == 0 {
					empty := "还没有节点"
					if strings.TrimSpace(a.search) != "" {
						empty = "没有匹配的节点"
					}
					ui.Text(c, empty).FontSize(fsCaption).TextColor(p.faint).Padding(4, 4)
				}
				lastStage := ""
				for _, n := range nodes {
					node := n
					if node.StageTitle != "" && node.StageTitle != lastStage {
						lastStage = node.StageTitle
						ui.Text(c, lastStage).FontSize(fsCaption).TextColor(p.faint).
							Padding(8, 4, 2, 4)
					}
					selected := node.ID == a.selectedNode().ID
					row := ui.ButtonBase(c.Key("node:" + node.ID)).FillWidth().Height(24).
						Radius(radiusBadge).PaddingX(5).Children(func() {
						ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
							a.nodeIndicator(c, t, node.Status)
							ui.Text(c, node.Title).FontSize(fsUI).FontWeight(weightFor(selected)).
								TextColor(p.sidebarText).MaxLines(1)
						})
					})
					if selected {
						row.Background(p.blueStrong)
					} else if row.Hovered() {
						row.Background(p.navHover)
					}
					if row.Clicked() {
						a.routeNodeID = node.ID
					}
				}
			})
		})
	})
}

func weightFor(selected bool) int {
	if selected {
		return 600
	}
	return 400
}

func (a *app) nodeIndicator(c *ui.Context, t *ui.Theme, status string) {
	p := paletteFor(t.Dark)
	slot := ui.Box(c).Size(12, 12).Center()
	switch status {
	case "done":
		slot.Children(func() {
			ui.Text(c, "✓").FontSize(12).FontWeight(600).TextColor(p.green)
		})
	case "current":
		slot.Children(func() {
			ui.Box(c).Size(7, 7).Radius(3.5).Background(p.green)
		})
	}
}
