package main

import (
	"fmt"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

// navPanel is the one place to move between learning routes and past
// conversations. It hangs from the small control in the top-right corner and
// never moves the reading column: the rail and the ability profile stay where
// they are.
func (a *app) navPanel(c *ui.Context, t *ui.Theme, p palette, panel, anchor ui.Element) {
	panel.AttachTo(anchor, ui.AnchorTopEnd, ui.AnchorBottomEnd).Margin(0, 0, 6, 0).
		Background(p.panelBg).BorderWidth(1).BorderColor(p.hairline).Radius(16).Padding(6).Width(304).
		Shadow(0, 18, 46, 0, withAlpha(ui.Hex("#000000"), 0.28)).
		Transition(ui.ElementTransition{
			Animation: &ui.Snappy,
			Enter:     &ui.Motion{Scale: 0.97, Y: 4},
			Exit:      &ui.Motion{Scale: 0.97, Y: 4},
		})
	panel.Children(func() {
		ui.Scroll(c).FillWidth().MaxHeight(460).Children(func() {
			ui.Column(c).FillWidth().Gap(1).Children(func() {
				a.navHeader(c, p, "路线", "新路线", func() { a.newRoute() })
				if len(a.goals) == 0 {
					a.navEmpty(c, p, "还没有路线。")
				}
				for _, goal := range a.goals {
					a.navRouteRow(c, p, goal)
				}

				ui.Box(c).FillWidth().Height(1).Background(p.hairline).Margin(8, 6)
				a.navHeader(c, p, "对话", "新对话", func() { a.newConversation() })
				groups := groupConversations(a.convs)
				if len(groups) == 0 {
					a.navEmpty(c, p, "还没有对话。")
				}
				for _, group := range groups {
					ui.Text(c, group.label).FontSize(9).Font(monoFont).TextColor(p.faint).
						Padding(8, 8, 3, 8)
					for _, conv := range group.items {
						a.navConversationRow(c, p, conv)
					}
				}
			})
		})
	})
}

func (a *app) navEmpty(c *ui.Context, p palette, text string) {
	ui.Text(c, text).FontSize(fsCaption).TextColor(p.faint).Padding(3, 8, 8, 8)
}

func (a *app) navHeader(c *ui.Context, p palette, title, action string, run func()) {
	ui.Row(c).FillWidth().Height(26).PaddingX(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontSize(9).Font(monoFont).TextColor(p.faint)
		ui.Spacer(c)
		btn := ui.ButtonBase(c).Height(22).PaddingX(6).Radius(radiusBadge).Label(action).
			Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
		btn.Children(func() {
			ui.Text(c, action).FontSize(9.5).TextColor(p.muted)
		})
		if btn.Hovered() {
			btn.Background(p.navHover)
		}
		if btn.Clicked() {
			run()
		}
	})
}

// navRow lays out one entry: the clickable body takes the width, and the
// delete control only appears while the row is under the pointer.
func (a *app) navRow(c *ui.Context, p palette, height float32, label string, body, open func(), delLabel string, del func()) {
	ui.Row(c).FillWidth().Height(height).AlignItems(ui.Center).Children(func() {
		main := ui.ButtonBase(c).Grow(1).MinWidth(0).Height(height).Radius(radiusControl).
			PaddingX(8).Justify(ui.Start).AlignItems(ui.Start).Label(label).
			Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
		main.Children(body)
		if main.Hovered() {
			main.Background(p.navHover)
		}
		if main.Clicked() {
			a.navOpen = false
			open()
		}

		remove := ui.ButtonBase(c).Size(24, 24).Radius(radiusBadge).Label(delLabel).
			Tooltip(delLabel).Center()
		if main.Hovered() || remove.Hovered() {
			remove.Children(func() {
				ui.Icon(c, iconTrash).Size(13, 13).TextColor(p.faint)
			})
			if remove.Hovered() {
				remove.Background(p.rustSoft)
			}
			if remove.Clicked() && del != nil {
				del()
			}
		}
	})
}

func (a *app) navRouteRow(c *ui.Context, p palette, goal store.Goal) {
	active := a.goal.ID == goal.ID
	a.navRow(c, p, 40, goal.Title, func() {
		ui.Column(c).FillWidth().Gap(1).Children(func() {
			ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, goal.Title).Grow(1).MinWidth(0).FontSize(fsUI).
					TextColor(p.text).MaxLines(1)
				if active {
					ui.Box(c).Size(4, 4).Radius(2).Background(p.accent)
				}
			})
			detail := fmt.Sprintf("%d/%d 节", goal.Done, goal.Total)
			if goal.CurrentTitle != "" {
				detail += " · " + goal.CurrentTitle
			}
			ui.Text(c, detail).FontSize(9.5).TextColor(p.faint).MaxLines(1)
		})
	}, func() { a.selectGoal(goal.ID) }, "删除路线", func() { a.askDeleteRoute(goal.ID) })
}

func (a *app) navConversationRow(c *ui.Context, p palette, conv store.Conversation) {
	active := a.active.ID == conv.ID
	a.navRow(c, p, 32, conv.Title, func() {
		ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, conv.Title).Grow(1).MinWidth(0).FontSize(fsUI).
				TextColor(p.body).MaxLines(1)
			if active {
				ui.Box(c).Size(4, 4).Radius(2).Background(p.accent)
			}
		})
	}, func() { a.selectConversation(conv.ID) }, "删除对话", func() { a.askDeleteConversation(conv.ID) })
}

// dayStart is the local midnight of t. Truncating the instant would use the
// UTC day and put the boundary eight hours off in Asia/Shanghai.
func dayStart(t time.Time) time.Time {
	t = t.Local()
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
