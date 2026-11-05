package main

import (
	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

var masteryLabels = []string{"未学", "初学", "熟练", "精通"}

func hairline(c *ui.Context, color ui.Color) {
	ui.Box(c).FillWidth().Height(1).Background(color)
}

func (a *app) routePanel(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	bar := c.TitleBar()
	ui.Column(c).Width(288).FillHeight().Background(p.surface2).
		BorderWidth(1).BorderColor(p.hairline).Radius(10).Children(func() {
		ui.Row(c).Height(max(bar.Height, 40)).Padding(0, bar.Right+12, 0, 12).
			Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "学习路线").FontSize(fsUI).FontWeight(600).TextColor(p.text)
			ui.Textf(c, "%d 节点", len(a.nodes)).FontSize(fsCaption).TextColor(p.faint)
			ui.Spacer(c)
			closeBtn := ui.ButtonBase(c).Size(20, 20).Radius(radiusBadge).Center().Children(func() {
				ui.Text(c, "×").FontSize(13).TextColor(p.muted)
			})
			if closeBtn.Hovered() {
				closeBtn.Background(p.fieldBg)
			}
			if closeBtn.Clicked() {
				a.routeOpen = false
			}
		})

		ui.Scroll(c).Fill().TrackScroll(&a.routeScroll).Children(func() {
			ui.Column(c).FillWidth().Padding(10, 8).Gap(8).Children(func() {
				// Routes
				ui.Column(c).FillWidth().Gap(1).Children(func() {
					for _, goal := range a.goals {
						g := goal
						active := g.ID == a.goal.ID
						weight := 400
						if active {
							weight = 600
						}
						row := ui.ButtonBase(c.Key("route:" + g.ID)).FillWidth().Height(22).Radius(radiusBadge).
							PaddingX(6).Children(func() {
							ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
								ui.Text(c, g.Title).FontSize(fsCaption).FontWeight(weight).MaxLines(1).TextColor(p.text)
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
					}
					newRoute := ui.ButtonBase(c).FillWidth().Height(22).Radius(radiusBadge).
						PaddingX(6).Justify(ui.Start).Children(func() {
						ui.Text(c, "＋ 新路线").FontSize(fsCaption).TextColor(p.faint)
					})
					if newRoute.Hovered() {
						newRoute.Background(p.navHover)
					}
					if newRoute.Clicked() {
						a.newRoute()
					}
				})

				hairline(c, p.hairline)

				node := a.selectedNode()
				if node.ID == "" {
					ui.Text(c, "还没有节点。在聊天里告诉教练你想学什么。").
						FontSize(fsCaption).TextColor(p.faint).LineHeight(1.5)
				} else {
					a.heatmapSection(c, t, node)
				}

				if len(a.nodes) > 0 {
					hairline(c, p.hairline)
					a.nodeTimeline(c, t)
				}

				ui.Box(c).Height(4)
			})
		})
	})
}

func (a *app) heatmapSection(c *ui.Context, t *ui.Theme, node store.Node) {
	p := paletteFor(t.Dark)
	lit := 0
	for _, cell := range node.Cells {
		if cell.Mastery >= node.Threshold {
			lit++
		}
	}
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, node.Title).FontSize(fsUI).FontWeight(600).TextColor(p.text).MaxLines(2)
			ui.Spacer(c)
			ui.Text(c, statusLabel(node.Status)).FontSize(fsCaption).TextColor(statusTint(node.Status, p))
		})
		if len(node.Cells) == 0 {
			ui.Text(c, "还没有热力图。开始这一节时教练会定义 3–9 个格子。").
				FontSize(fsCaption).TextColor(p.faint).LineHeight(1.5)
			return
		}
		ui.Textf(c, "%d / %d 格达到%s", lit, len(node.Cells), masteryLabels[node.Threshold]).
			FontSize(fsCaption).TextColor(p.faint)
		ui.Grid(c).FillWidth().Columns(3).Gap(4).Children(func() {
			for _, skill := range node.Cells {
				cell := skill
				bg, border, fg := heatStyle(cell.Mastery, t.Dark)
				ui.Box(c).Height(50).Radius(radiusBadge).Background(bg).BorderWidth(1).BorderColor(border).
					Padding(5).Children(func() {
					ui.Text(c, cell.Title).FontSize(fsCaption).FontWeight(600).TextColor(fg).
						MaxLines(2).LineHeight(1.3)
				})
			}
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			for level := 0; level < 4; level++ {
				bg, border, _ := heatStyle(level, t.Dark)
				ui.Row(c).Gap(3).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(6, 6).Radius(1).Background(bg).BorderWidth(1).BorderColor(border)
					ui.Text(c, masteryLabels[level]).FontSize(fsCaption).TextColor(p.faint)
				})
			}
		})
	})
}

func (a *app) nodeTimeline(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	selectedID := a.selectedNode().ID
	lastStage := ""
	ui.Column(c).FillWidth().Gap(0).Children(func() {
		for _, n := range a.nodes {
			node := n
			if node.StageTitle != "" && node.StageTitle != lastStage {
				lastStage = node.StageTitle
				ui.Text(c, lastStage).FontSize(fsCaption).TextColor(p.faint).
					LetterSpacing(0.5).Padding(6, 4, 1, 4)
			}
			selected := node.ID == selectedID
			row := ui.ButtonBase(c.Key("node:" + node.ID)).FillWidth().Height(22).Radius(radiusBadge).
				PaddingX(5).Children(func() {
				ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
					a.nodeDot(c, t, node.Status)
					ui.Text(c, node.Title).FontSize(fsCaption).FontWeight(weightFor(selected)).
						TextColor(p.text).MaxLines(1)
				})
			})
			if selected {
				row.Background(p.navActive)
			} else if row.Hovered() {
				row.Background(p.navHover)
			}
			if row.Clicked() {
				a.routeNodeID = node.ID
			}
		}
	})
}

func weightFor(selected bool) int {
	if selected {
		return 600
	}
	return 400
}

func (a *app) nodeDot(c *ui.Context, t *ui.Theme, status string) {
	p := paletteFor(t.Dark)
	dot := ui.Box(c).Size(6, 6).Radius(3)
	switch status {
	case "done":
		dot.Background(p.blue)
	case "current":
		dot.Background(p.blueSoft).BorderWidth(1).BorderColor(p.blue)
	default:
		dot.Background(p.canvas).BorderWidth(1).BorderColor(p.borderStrong)
	}
}

func statusLabel(status string) string {
	switch status {
	case "done":
		return "掌握"
	case "current":
		return "学习中"
	default:
		return "待学"
	}
}

func statusTint(status string, p palette) ui.Color {
	switch status {
	case "current":
		return p.blue
	case "done":
		return p.faint
	default:
		return p.faint
	}
}
