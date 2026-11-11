package main

import (
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

const (
	heatRailRight   float32 = 12
	heatRailWidth   float32 = 54
	heatRowHeight   float32 = 22
	heatRowGap      float32 = 7
	heatPipSize     float32 = 5
	heatPipGap      float32 = 2
	heatMaxPips             = 7
	heatRestOpacity float32 = 0.62
	heatCardWidth   float32 = 220
	heatCardRight   float32 = 78
	heatCardHeight  float32 = 152
)

var (
	heatColor = ui.Animation{Duration: 160 * time.Millisecond, Ease: ui.EaseOut}
	heatFade  = ui.Animation{Duration: 220 * time.Millisecond, Ease: ui.EaseOut}
)

// abilityHeat is one transferable ability, not a route node. Cells from any
// node in the route contribute to the same ability.
type abilityHeat struct {
	Ability  store.Ability
	Cells    []store.Cell
	Nodes    int
	Mastered int
	Sum      int
	Average  float32
	Level    int
	Counts   [4]int
}

func abilityHeats(abilities []store.Ability, nodes []store.Node) []abilityHeat {
	heats := make([]abilityHeat, len(abilities))
	byID := make(map[string]int, len(abilities))
	for i, ability := range abilities {
		heats[i].Ability = ability
		byID[ability.ID] = i
	}
	seenNodes := make([]map[string]bool, len(abilities))
	for i := range seenNodes {
		seenNodes[i] = map[string]bool{}
	}
	for _, node := range nodes {
		for _, cell := range node.Cells {
			index, ok := byID[cell.AbilityID]
			if !ok {
				continue
			}
			heats[index].Cells = append(heats[index].Cells, cell)
			heats[index].Sum += cell.Mastery
			if cell.Mastery >= 0 && cell.Mastery <= 3 {
				heats[index].Counts[cell.Mastery]++
			}
			if cell.Mastery >= 3 {
				heats[index].Mastered++
			}
			if !seenNodes[index][node.ID] {
				seenNodes[index][node.ID] = true
				heats[index].Nodes++
			}
		}
	}
	for i := range heats {
		if len(heats[i].Cells) == 0 {
			continue
		}
		heats[i].Average = float32(heats[i].Sum) / float32(len(heats[i].Cells))
		switch {
		case heats[i].Average >= 2.5:
			heats[i].Level = 3
		case heats[i].Average >= 1.5:
			heats[i].Level = 2
		case heats[i].Average > 0:
			heats[i].Level = 1
		}
	}
	return heats
}

func masteryLabel(mastery int) string {
	switch mastery {
	case 1:
		return "初学"
	case 2:
		return "熟练"
	case 3:
		return "精通"
	default:
		return "未掌握"
	}
}

func abilityLevelLabel(level int) string {
	switch level {
	case 1:
		return "初学"
	case 2:
		return "熟练"
	case 3:
		return "精通"
	default:
		return "未开始"
	}
}

// heatPanel is a quiet ability profile in the right margin. It is one map
// for the whole route: each row is a transferable ability, and every small
// block is one knowledge cell that supplies evidence for it. No labels or
// counts sit on the canvas; the names and detail live only in the hover card.
// Rows never expand or overlap; the panel fades back at rest.
func (a *app) heatPanel(c *ui.Context, t *ui.Theme) {
	if len(a.abilities) == 0 || a.home || a.panelOpen || a.settingsOpen {
		return
	}
	p := paletteFor(t.Dark)
	heats := abilityHeats(a.abilities, a.nodes)
	var panel ui.Element
	hovered := -1
	panelHovered := false
	ui.Column(c).Absolute().Right(heatRailRight).Top(0).Bottom(0).Width(heatRailWidth).
		Justify(ui.Center).AlignItems(ui.End).PassThrough().Children(func() {
		panel = ui.Column(c).Width(heatRailWidth).AlignItems(ui.End).
			Role(ui.RoleGroup).Label("能力热力图")
		panel.Children(func() {
			ui.Column(c).Width(heatRailWidth).AlignItems(ui.End).Gap(heatRowGap).Children(func() {
				for i := range heats {
					if a.heatAbilityRow(c, t, p, heats[i]) {
						hovered = i
					}
				}
			})
		})
		panelHovered = panel.Hovered()
	})
	a.heatHover = hovered
	target := heatRestOpacity
	if panelHovered {
		target = 1
	}
	shown := target
	if !c.Preferences().ReduceMotion {
		shown = panel.AnimateTo("heat-fade", target, heatFade)
	}
	a.heatShown = shown
	panel.Opacity(shown)
	a.heatCard(c, t, p, heats)
}

func (a *app) heatAbilityRow(c *ui.Context, t *ui.Theme, p palette, heat abilityHeat) bool {
	row := ui.Row(c).Width(heatRailWidth).Height(heatRowHeight).AlignItems(ui.Center).
		Justify(ui.End).Label(heat.Ability.Title)
	row.Children(func() {
		visible := heat.Cells
		extra := 0
		if len(visible) > heatMaxPips {
			extra = len(visible) - heatMaxPips
			visible = visible[:heatMaxPips]
		}
		pips := ui.Row(c).Gap(heatPipGap).AlignItems(ui.Center)
		pips.Children(func() {
			if len(visible) == 0 {
				ui.Box(c).Size(heatPipSize, heatPipSize).Radius(1.5).
					Background(p.surface).BorderWidth(1).BorderColor(p.hairline)
				return
			}
			for _, cell := range visible {
				fill, border, _ := heatStyle(cell.Mastery, t.Dark)
				ui.Box(c).Size(heatPipSize, heatPipSize).Radius(1.5).
					Background(fill).BorderWidth(1).BorderColor(border)
			}
			if extra > 0 {
				ui.Textf(c, "+%d", extra).FontSize(8).Font(monoFont).TextColor(p.faint)
			}
		})
	})
	hovered := row.Hovered()
	if hovered {
		row.Background(p.surface).Radius(6)
	}
	return hovered
}

// heatCard is one fixed detail surface for the whole profile. Its position
// and size never move; moving between ability rows only changes its text.
func (a *app) heatCard(c *ui.Context, t *ui.Theme, p palette, heats []abilityHeat) {
	if len(heats) == 0 {
		return
	}
	index := a.heatHover
	if index < 0 || index >= len(heats) {
		index = 0
	}
	heat := heats[index]
	show := a.heatHover >= 0
	ui.Column(c).Absolute().Right(heatCardRight).Top(0).Bottom(0).Width(heatCardWidth).
		Justify(ui.Center).PassThrough().Children(func() {
		card := ui.Column(c).Width(heatCardWidth).Height(heatCardHeight).
			Padding(12, 16).Radius(16).
			Material(glass.Glass{Style: glass.Regular, Clarity: 0.32, Tint: withAlpha(p.panelBg, 0.92)}).
			BorderWidth(1).BorderColor(p.hairline).Shadow(0, 12, 28, 0, withAlpha(p.shadow, 0.8)).
			Label("能力详情").PassThrough()
		card.Children(func() {
			ui.Column(c).FillWidth().Gap(3).Children(func() {
				ui.Textf(c, "能力 %s / %s", twoDigits(index+1), twoDigits(len(heats))).
					FontSize(9).Font(monoFont).TextColor(p.faint)
				ui.Text(c, heat.Ability.Title).FontSize(14).FontWeight(400).
					TextColor(p.text).MaxLines(1).Padding(2, 0, 0, 0)
				if heat.Ability.Description != "" {
					ui.Text(c, heat.Ability.Description).FontSize(10.5).LineHeight(1.5).
						TextColor(p.muted).MaxLines(2)
				} else {
					ui.Text(c, "还没有写下这项能力的具体含义。").FontSize(10.5).
						LineHeight(1.5).TextColor(p.faint).MaxLines(2)
				}
				ui.Box(c).FillWidth().Height(2).Radius(1).Background(p.fieldBg).Clip().
					Margin(6, 0, 3, 0).Children(func() {
					percent := int(heat.Average*100/3 + 0.5)
					if percent > 0 {
						ui.Box(c).WidthPercent(float32(percent)).Height(2).Radius(1).Background(p.accent)
					}
				})
				ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
					ui.Text(c, abilityLevelLabel(heat.Level)).FontSize(10.5).TextColor(p.accent)
					ui.Spacer(c)
					if len(heat.Cells) > 0 {
						ui.Textf(c, "%d%%", int(heat.Average*100/3+0.5)).
							FontSize(10).Font(monoFont).TextColor(p.muted)
					}
				})
				ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
					shown := 0
					for level := 1; level <= 3; level++ {
						if heat.Counts[level] == 0 {
							continue
						}
						shown++
						fill, _, _ := heatStyle(level, t.Dark)
						ui.Box(c).Size(4, 4).Radius(2).Background(fill)
						ui.Textf(c, "%d %s", heat.Counts[level], masteryLabel(level)).
							FontSize(9).TextColor(p.muted)
					}
					if shown == 0 {
						ui.Text(c, "全部待测").FontSize(9).TextColor(p.faint)
					}
				})
				if len(heat.Cells) == 0 {
					ui.Text(c, "还没有格子归属这项能力").FontSize(9.5).TextColor(p.faint)
				} else {
					ui.Textf(c, "覆盖 %d 节 · %d 个格子", heat.Nodes, len(heat.Cells)).
						FontSize(9.5).TextColor(p.faint)
				}
			})
		})
		target := float32(0)
		if show {
			target = 1
		}
		opacity := target
		if !c.Preferences().ReduceMotion {
			opacity = card.AnimateTo("heat-card", target, heatFade)
		}
		a.heatCardShown = opacity
		card.Opacity(opacity)
	})
}
