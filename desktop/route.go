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
	ui.Box(c).Width(304).FillHeight().Padding(8, 8, 8, 0).Children(func() {
		ui.Column(c).Fill().FillHeight().Background(p.surface).
			BorderWidth(1).BorderColor(p.hairline).Radius(10).Clip().Children(func() {
			ui.Row(c).Height(max(bar.Height, 40)).Padding(0, bar.Right+12, 0, 12).
				Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "知识热力图").FontSize(fsUI).FontWeight(600).TextColor(p.text)
				ui.Spacer(c)
				if a.current.ID != "" {
					ui.Text(c, statusLabel(a.current.Status)).FontSize(fsCaption).
						TextColor(statusTint(a.current.Status, p))
				}
			})
			ui.Scroll(c).Fill().TrackScroll(&a.routeScroll).Children(func() {
				ui.Column(c).FillWidth().Padding(12, 10).Gap(10).Children(func() {
					if a.current.ID == "" {
						ui.Text(c, "当前没有节点。在聊天里开始学习后，这里会显示本节的热力图。").
							FontSize(fsCaption).TextColor(p.faint).LineHeight(1.55)
						return
					}
					a.heatmapSection(c, t, a.current)
				})
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
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		ui.Text(c, node.Title).FontSize(fsTitle).FontWeight(600).TextColor(p.text).MaxLines(2)
		if len(node.Cells) == 0 {
			ui.Text(c, "还没有热力图。开始这一节时教练会定义 3–9 个格子。").
				FontSize(fsCaption).TextColor(p.faint).LineHeight(1.55)
			return
		}
		ui.Textf(c, "%d / %d 格达到%s", lit, len(node.Cells), masteryLabels[node.Threshold]).
			FontSize(fsCaption).TextColor(p.faint)
		ui.Grid(c).FillWidth().Columns(3).Gap(6).Children(func() {
			for _, skill := range node.Cells {
				cell := skill
				bg, border, fg := heatStyle(cell.Mastery, t.Dark)
				ui.Box(c).Height(64).Radius(radiusBadge).Background(bg).BorderWidth(1).BorderColor(border).
					Padding(6).Children(func() {
					ui.Column(c).FillWidth().FillHeight().Gap(2).Children(func() {
						ui.Text(c, cell.Title).FontSize(fsCaption).FontWeight(600).TextColor(fg).
							MaxLines(2).LineHeight(1.3)
						ui.Spacer(c)
						ui.Text(c, masteryLabels[cell.Mastery]).FontSize(fsCaption).TextColor(fg)
					})
				})
			}
		})
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			for level := 0; level < 4; level++ {
				bg, border, _ := heatStyle(level, t.Dark)
				ui.Row(c).Gap(3).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(7, 7).Radius(1).Background(bg).BorderWidth(1).BorderColor(border)
					ui.Text(c, masteryLabels[level]).FontSize(fsCaption).TextColor(p.faint)
				})
			}
		})
	})
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
	if status == "current" {
		return p.blue
	}
	return p.faint
}
