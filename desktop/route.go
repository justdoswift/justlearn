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
				Gap(6).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(p.hairline).Children(func() {
				ui.Text(c, "知识热力图").FontSize(fsUI).FontWeight(600).TextColor(p.text)
				ui.Spacer(c)
				if node := a.selectedNode(); node.ID != "" {
					a.nodeIndicator(c, t, node.Status)
				}
			})
			ui.Scroll(c).Fill().TrackScroll(&a.routeScroll).Children(func() {
				ui.Column(c).FillWidth().Padding(12, 10).Gap(10).Children(func() {
					node := a.selectedNode()
					if node.ID == "" {
						ui.Text(c, "当前没有节点。在聊天里开始学习后，这里会显示本节的热力图。").
							FontSize(fsCaption).TextColor(p.faint).LineHeight(1.55)
						return
					}
					a.heatmapSection(c, t, node)
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
		ui.Text(c, node.Title).FontSize(fsSection).FontWeight(600).TextColor(p.text).MaxLines(2)
		if len(node.Cells) == 0 {
			ui.Text(c, "还没有热力图。开始这一节时教练会定义 3–9 个格子。").
				FontSize(fsCaption).TextColor(p.faint).LineHeight(1.55)
			return
		}
		ui.Textf(c, "%d / %d 格达到%s", lit, len(node.Cells), masteryLabels[node.Threshold]).
			FontSize(fsUI).TextColor(p.faint)
		ui.Grid(c).FillWidth().Columns(7).Gap(4).Children(func() {
			for _, skill := range node.Cells {
				cell := skill
				bg, border, _ := heatStyle(cell.Mastery, t.Dark)
				ui.Box(c).Size(30, 30).Radius(4).Background(bg).BorderWidth(1).BorderColor(border).
					Tooltip(cell.Title + " · " + masteryLabels[cell.Mastery])
			}
		})
		hairline(c, p.hairline)
		ui.Text(c, "格子").FontSize(fsCaption).TextColor(p.faint)
		ui.Column(c).FillWidth().Gap(3).Children(func() {
			for _, skill := range node.Cells {
				cell := skill
				bg, border, _ := heatStyle(cell.Mastery, t.Dark)
				ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(10, 10).Radius(2).Background(bg).BorderWidth(1).BorderColor(border)
					ui.Text(c, cell.Title).FontSize(fsUI).TextColor(p.body).Grow(1).MaxLines(1)
					ui.Text(c, masteryLabels[cell.Mastery]).FontSize(fsCaption).TextColor(p.faint)
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
