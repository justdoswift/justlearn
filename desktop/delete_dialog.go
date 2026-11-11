package main

import (
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

func (a *app) deleteDialogs(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	if a.confirmRouteOpen {
		title := "这条路线"
		for _, g := range a.goals {
			if g.ID == a.pendingRouteID {
				title = g.Title
				break
			}
		}
		ui.DialogBase(c, &a.confirmRouteOpen, func(backdrop, panel ui.Element) {
			backdrop.Background(ui.RGBA(0, 0, 0, 0.32))
			panel.Background(p.panelBg).BorderWidth(1).BorderColor(p.hairline).Radius(20).
				Shadow(0, 20, 56, 0, withAlpha(ui.Hex("#000000"), 0.32)).
				Padding(20).Width(340).Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.96, Y: 6}}).Children(func() {
				ui.Column(c).FillWidth().Gap(10).Children(func() {
					ui.Text(c, "删除这条路线？").FontSize(fsTitle).FontWeight(400).TextColor(p.text)
					ui.Textf(c, "「%s」的节点、格子、相关会话和消息都会删除，不能撤销。", title).
						FontSize(fsCaption).TextColor(p.muted).LineHeight(1.55)
					ui.Row(c).FillWidth().Gap(6).Justify(ui.End).Children(func() {
						cancel := ui.ButtonBase(c).Width(72).Height(30).Radius(radiusControl).Center().
							Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond}).Children(func() {
							ui.Text(c, "取消").FontSize(fsUI).TextColor(p.muted)
						})
						if cancel.Hovered() {
							cancel.Background(p.navHover)
						}
						if cancel.Clicked() {
							a.confirmRouteOpen = false
							a.pendingRouteID = ""
						}
						del := ui.ButtonBase(c).Width(72).Height(30).Radius(radiusControl).
							Material(glass.Glass{Tint: p.rust, Interactive: true}).Center().Children(func() {
							ui.Text(c, "删除").FontSize(fsUI).FontWeight(400).TextColor(ui.Hex("#FFFFFF"))
						})
						if del.Clicked() {
							a.deleteRoute(a.pendingRouteID)
						}
					})
				})
			})
		})
	}
	if a.confirmChatOpen {
		title := "这个对话"
		for _, conv := range a.convs {
			if conv.ID == a.pendingChatID {
				title = conv.Title
				break
			}
		}
		ui.DialogBase(c, &a.confirmChatOpen, func(backdrop, panel ui.Element) {
			backdrop.Background(ui.RGBA(0, 0, 0, 0.32))
			panel.Background(p.panelBg).BorderWidth(1).BorderColor(p.hairline).Radius(20).
				Shadow(0, 20, 56, 0, withAlpha(ui.Hex("#000000"), 0.32)).
				Padding(20).Width(340).Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.96, Y: 6}}).Children(func() {
				ui.Column(c).FillWidth().Gap(10).Children(func() {
					ui.Text(c, "删除这个对话？").FontSize(fsTitle).FontWeight(400).TextColor(p.text)
					ui.Textf(c, "「%s」的消息会一起删除，学习路线保留。", title).
						FontSize(fsCaption).TextColor(p.muted).LineHeight(1.55)
					ui.Row(c).FillWidth().Gap(6).Justify(ui.End).Children(func() {
						cancel := ui.ButtonBase(c).Width(72).Height(30).Radius(radiusControl).Center().
							Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond}).Children(func() {
							ui.Text(c, "取消").FontSize(fsUI).TextColor(p.muted)
						})
						if cancel.Hovered() {
							cancel.Background(p.navHover)
						}
						if cancel.Clicked() {
							a.confirmChatOpen = false
							a.pendingChatID = ""
						}
						del := ui.ButtonBase(c).Width(72).Height(30).Radius(radiusControl).
							Material(glass.Glass{Tint: p.rust, Interactive: true}).Center().Children(func() {
							ui.Text(c, "删除").FontSize(fsUI).FontWeight(400).TextColor(ui.Hex("#FFFFFF"))
						})
						if del.Clicked() {
							a.deleteConversation(a.pendingChatID)
						}
					})
				})
			})
		})
	}
}
