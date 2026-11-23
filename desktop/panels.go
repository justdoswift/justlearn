package main

import (
	"strings"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

const (
	railLeft      float32 = 18
	railHitWidth  float32 = 56
	railItemSize  float32 = 28
	railTickWidth float32 = 26
	railTickH     float32 = 4
	railCardWidth float32 = 296
	railCardGap   float32 = 10
)

var (
	panelSpring = ui.Animation{Spring: ui.SpringOptions{Duration: 320 * time.Millisecond, Bounce: 0.05}}
	railSpring  = ui.Animation{Spring: ui.SpringOptions{Duration: 220 * time.Millisecond, Bounce: 0.08}}
)

type routeStage struct {
	Title string
	Nodes []store.Node
}

func routeStages(nodes []store.Node) []routeStage {
	var out []routeStage
	for _, node := range nodes {
		title := strings.TrimSpace(node.StageTitle)
		if title == "" {
			title = "未分组"
		}
		if len(out) == 0 || out[len(out)-1].Title != title {
			out = append(out, routeStage{Title: title})
		}
		out[len(out)-1].Nodes = append(out[len(out)-1].Nodes, node)
	}
	return out
}

func routeProgress(nodes []store.Node) (done, total int) {
	total = len(nodes)
	for _, node := range nodes {
		if node.Status == "done" {
			done++
		}
	}
	return done, total
}

func (a *app) stageIsCurrent(stage routeStage) bool {
	for _, node := range stage.Nodes {
		if node.Status == "current" || a.current.ID != "" && node.ID == a.current.ID {
			return true
		}
	}
	return false
}

// railStageIndex is the stage the rail rests on: the current stage, the
// next unfinished one, or the first. A stage picked on the rail wins.
func (a *app) railStageIndex(stages []routeStage) int {
	if a.railSelected >= 0 && a.railSelected < len(stages) {
		return a.railSelected
	}
	for i, stage := range stages {
		if a.stageIsCurrent(stage) {
			return i
		}
	}
	for i, stage := range stages {
		done, total := routeProgress(stage.Nodes)
		if done < total {
			return i
		}
	}
	return 0
}

func railTickScale(distance int) float32 {
	switch distance {
	case 0:
		return 1
	case 1:
		return 0.74
	case 2:
		return 0.55
	default:
		return 0.4
	}
}

// routeRail is the roadmap as a quiet scale in the left margin. Each stage is
// one tick: a faint track with the finished part filled, so progress reads
// without a single label. Hovering a tick floats its preview beside it;
// clicking pins that preview so it stays while the pointer moves.
func (a *app) routeRail(c *ui.Context, t *ui.Theme) {
	stages := routeStages(a.nodes)
	if len(stages) == 0 {
		return
	}
	p := paletteFor(t.Dark)
	selected := a.railStageIndex(stages)
	hovered := -1
	buttons := make([]ui.Element, len(stages))
	tracks := make([]ui.Element, len(stages))
	fills := make([]ui.Element, len(stages))

	ui.Overlay(c, func() {
		ui.Column(c).Absolute().Left(railLeft).Top(0).Bottom(0).Width(railHitWidth).
			Justify(ui.Center).AlignItems(ui.Start).PassThrough().Children(func() {
			for i, stage := range stages {
				i, stage := i, stage
				button := ui.ButtonBase(c).Width(railHitWidth).Height(railItemSize).
					Radius(radiusBadge).Justify(ui.Start).KeepFocus().FocusRing(false).
					Label(railStageLabel(i, stage)).Description(stage.Title)
				button.Children(func() {
					// The stage number is what turns a column of dashes into
					// a scale. Two ticks alone read as a rendering accident.
					ui.Text(c, twoDigits(i+1)).Width(16).FontSize(9).Font(monoFont).
						TextColor(withAlpha(p.faint, 0.9))
					tracks[i] = ui.Box(c).Width(railTickWidth).Height(railTickH).Radius(railTickH / 2).
						Background(withAlpha(p.faint, 0.34)).TransformOrigin(0, 0.5).Clip()
					tracks[i].Children(func() {
						fills[i] = ui.Box(c).Height(railTickH).Radius(railTickH / 2).
							TransformOrigin(0, 0.5)
					})
				})
				buttons[i] = button
			}
		})

		for i := range buttons {
			if buttons[i].Hovered() || buttons[i].Focused() {
				hovered = i
			}
		}
		for i := range buttons {
			distance := i - selected
			if hovered >= 0 {
				distance = i - hovered
			}
			if distance < 0 {
				distance = -distance
			}
			done, total := routeProgress(stages[i].Nodes)
			percent := float32(0)
			if total > 0 {
				percent = float32(done) / float32(total)
			}
			if percent > 0 && percent < 0.06 {
				percent = 0.06
			}
			fills[i].WidthPercent(percent * 100).
				Background(a.railTickColor(p, stages[i], i == selected))
			tracks[i].AnimatePose(ui.Pose{Scale: railTickScale(distance)}, railSpring)
			if buttons[i].Clicked() {
				if a.railSelected == i {
					a.railSelected = -1
				} else {
					a.railSelected = i
				}
				a.focusDraft = false
			}
		}

		// The card shows for the hovered stage, or stays pinned on the one
		// picked from the rail. Clicking a stage used to change nothing
		// visible, which made the whole rail feel dead.
		detail := selected
		anchor := buttons[selected]
		shown := float32(0)
		if hovered >= 0 {
			detail = hovered
			anchor = buttons[hovered]
			shown = 1
		} else if a.railSelected >= 0 {
			shown = 1
		}
		card := a.routeRailCard(c, t, p, stages[detail], detail, len(stages))
		card.AttachTo(anchor, ui.AnchorRight, ui.AnchorLeft).Margin(0, 0, 0, railCardGap)
		fade := card.AnimateTo("route-rail-card", shown, railSpring)
		a.railShown = fade
		card.Opacity(fade).Translate(-6*(1-fade), 0).Scale(0.985+0.015*fade, 0.985+0.015*fade).
			PassThrough()
	})
}

func railStageLabel(index int, stage routeStage) string {
	return "阶段 " + twoDigits(index+1) + "，" + stage.Title
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func (a *app) railTickColor(p palette, stage routeStage, selected bool) ui.Color {
	done, total := routeProgress(stage.Nodes)
	switch {
	case a.stageIsCurrent(stage):
		return p.accent
	case total > 0 && done == total:
		return withAlpha(p.accent, 0.55)
	case selected:
		return withAlpha(p.accent, 0.72)
	default:
		return withAlpha(p.faint, 0.72)
	}
}

func (a *app) routeRailCard(c *ui.Context, t *ui.Theme, p palette, stage routeStage, index, count int) ui.Element {
	done, total := routeProgress(stage.Nodes)
	percent := 0
	if total > 0 {
		percent = done * 100 / total
	}
	card := ui.Column(c).Width(railCardWidth).Padding(13, 15, 14, 15).Radius(16).
		Material(glass.Glass{Style: glass.Regular, Clarity: 0.35, Tint: withAlpha(p.composerTint, 0.9)}).
		BorderWidth(1).BorderColor(withAlpha(p.borderStrong, 0.45)).
		Shadow(0, 18, 44, 0, withAlpha(p.shadow, 0.8))
	card.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "阶段 %s", twoDigits(index+1)).FontSize(9).Font(monoFont).TextColor(p.faint)
			ui.Spacer(c)
			ui.Textf(c, "%d/%d", done, total).FontSize(9).Font(monoFont).TextColor(p.faint)
		})
		ui.Text(c, stage.Title).FontSize(14).FontWeight(400).LineHeight(1.35).TextColor(p.text).
			MaxLines(2).Padding(5, 0, 0, 0)
		ui.Box(c).FillWidth().Height(2).Radius(1).Background(p.fieldBg).Clip().Margin(8, 0, 10, 0).Children(func() {
			if percent > 0 {
				ui.Box(c).WidthPercent(float32(percent)).Height(2).Radius(1).Background(p.accent)
			}
		})
		visible := stage.Nodes
		if len(visible) > 5 {
			visible = visible[:5]
		}
		for _, node := range visible {
			a.routeRailNode(c, p, node)
		}
		if len(stage.Nodes) > len(visible) {
			ui.Textf(c, "另有 %d 个节点", len(stage.Nodes)-len(visible)).
				FontSize(9).TextColor(p.faint).Padding(2, 0, 0, 0)
		}
	})
	return card
}

func (a *app) routeRailNode(c *ui.Context, p palette, node store.Node) {
	current := node.Status == "current" || a.current.ID != "" && node.ID == a.current.ID
	done := node.Status == "done"
	fg := p.body
	if done {
		fg = p.muted
	}
	if current {
		fg = p.text
	}
	ui.Row(c).FillWidth().MinHeight(20).Gap(8).AlignItems(ui.Center).Children(func() {
		dot := ui.Box(c).Size(5, 5).Radius(2.5)
		switch {
		case current:
			dot.Background(p.accent)
		case done:
			dot.Background(withAlpha(p.accent, 0.55))
		default:
			dot.Background(p.panelBg).BorderWidth(1).BorderColor(p.borderStrong)
		}
		ui.Text(c, node.Title).Grow(1).MinWidth(0).FontSize(11.5).FontWeight(400).
			TextColor(fg).MaxLines(1)
		if current {
			ui.Text(c, "现在").FontSize(8.5).TextColor(p.accent)
		}
	})
}
