package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/justdoswift/justlearn/desktop/internal/genui"
	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

// canvasSection renders the structured messages the tutor generates. The
// model sends a typed payload, never HTML; the blocks use the same reading
// column, type scale and surfaces as the rest of the chat.
func (a *app) canvasSection(c *ui.Context, t *ui.Theme, msg store.Message) {
	p := paletteFor(t.Dark)
	var payload genui.Payload
	if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil || len(payload.Blocks) == 0 {
		markdown.View(c, msg.Text, a.mdOptions(t)).FillWidth()
		return
	}
	ui.Column(c).FillWidth().Gap(12).Children(func() {
		if payload.Title != "" {
			ui.Text(c, payload.Title).FontSize(fsTitle).FontWeight(400).TextColor(p.text)
		}
		for i, block := range payload.Blocks {
			a.canvasBlock(c, t, p, msg, i, block)
		}
	})
}

func (a *app) canvasBlock(c *ui.Context, t *ui.Theme, p palette, msg store.Message, index int, block genui.Block) {
	switch block.Type {
	case "text":
		markdown.View(c, block.Text, a.mdOptions(t)).FillWidth()
	case "callout":
		a.canvasCallout(c, t, p, block)
	case "quote":
		a.canvasQuote(c, t, p, block)
	case "steps":
		a.canvasSteps(c, p, block)
	case "table":
		a.canvasTable(c, p, block)
	case "compare":
		a.canvasCompare(c, p, block)
	case "code":
		a.codeBlock(c, p, fmt.Sprintf("%s:code:%d", msg.ID, index), block.Language, block.Code, block.Caption)
	case "flashcards":
		a.canvasFlashcards(c, p, msg, index, block)
	case "bar":
		a.canvasBars(c, p, block)
	case "divider":
		ui.Box(c).FillWidth().Height(1).Background(p.hairline).Margin(2, 0)
	}
}

func (a *app) canvasCallout(c *ui.Context, t *ui.Theme, p palette, block genui.Block) {
	fg, bg, line := p.accent, p.accentSoft, p.accentLine
	switch block.Tone {
	case "success":
		fg, bg, line = p.green, p.greenSoft, withAlpha(p.green, 0.35)
	case "warning":
		fg, bg, line = p.rust, p.rustSoft, withAlpha(p.rust, 0.35)
	}
	ui.Box(c).FillWidth().Background(bg).Radius(radiusControl).
		BorderWidth(0, 0, 0, 2).BorderColor(line).Padding(10, 12).Children(func() {
		ui.Column(c).FillWidth().Gap(4).Children(func() {
			if block.Title != "" {
				ui.Text(c, block.Title).FontSize(fsCaption).FontWeight(400).TextColor(fg)
			}
			markdown.View(c, block.Body, a.mdOptions(t)).FillWidth()
		})
	})
}

func (a *app) canvasQuote(c *ui.Context, t *ui.Theme, p palette, block genui.Block) {
	ui.Box(c).FillWidth().BorderWidth(0, 0, 0, 3).BorderColor(p.accent).Padding(2, 0, 2, 14).Children(func() {
		ui.Column(c).FillWidth().Gap(4).Children(func() {
			markdown.View(c, block.Text, a.mdOptions(t)).FillWidth()
			if block.Caption != "" {
				ui.Text(c, "— "+block.Caption).FontSize(fsCaption).TextColor(p.muted)
			}
		})
	})
}

func (a *app) canvasSteps(c *ui.Context, p palette, block genui.Block) {
	ui.Column(c).FillWidth().Gap(0).Children(func() {
		if block.Title != "" {
			ui.Text(c, block.Title).FontSize(fsUI).FontWeight(400).TextColor(p.text).Padding(0, 0, 8, 0)
		}
		for i, item := range block.Items {
			ui.Row(c).FillWidth().Gap(10).Padding(0, 0, 10, 0).AlignItems(ui.Start).Children(func() {
				ui.Box(c).Size(22, 22).Radius(11).Background(p.accentSoft).Center().Children(func() {
					ui.Text(c, strconv.Itoa(i+1)).FontSize(fsCaption).Font(monoFont).TextColor(p.accent)
				})
				ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
					if item.Title != "" {
						ui.Text(c, item.Title).FontSize(fsRead).FontWeight(400).TextColor(p.text)
					}
					if item.Body != "" {
						ui.Text(c, item.Body).FontSize(fsUI).TextColor(p.muted).LineHeight(1.55)
					}
				})
			})
			if i < len(block.Items)-1 {
				ui.Box(c).FillWidth().Height(1).Background(p.hairline).Margin(0, 0, 10, 32)
			}
		}
	})
}

func (a *app) canvasTable(c *ui.Context, p palette, block genui.Block) {
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		if block.Title != "" {
			ui.Text(c, block.Title).FontSize(fsUI).FontWeight(400).TextColor(p.text)
		}
		ui.Column(c).FillWidth().Radius(radiusCard).Background(p.surface).
			BorderWidth(1).BorderColor(p.hairline).Clip().Children(func() {
			ui.Row(c).FillWidth().Background(p.surface2).Padding(7, 10).Children(func() {
				for _, header := range block.Columns {
					ui.Text(c, header).Grow(1).MinWidth(0).FontSize(fsCaption).FontWeight(400).
						TextColor(p.muted).MaxLines(2)
				}
			})
			for ri, row := range block.Rows {
				ui.Row(c).FillWidth().Padding(7, 10).Children(func() {
					for i := range block.Columns {
						cell := ""
						if i < len(row) {
							cell = row[i]
						}
						ui.Text(c, cell).Grow(1).MinWidth(0).FontSize(fsUI).LineHeight(1.5).
							TextColor(p.body).MaxLines(4)
					}
				})
				if ri < len(block.Rows)-1 {
					ui.Box(c).FillWidth().Height(1).Background(p.hairline)
				}
			}
			if block.Caption != "" {
				ui.Text(c, block.Caption).FontSize(fsCaption).TextColor(p.faint).Padding(7, 10, 8, 10)
			}
		})
	})
}

func (a *app) canvasCompare(c *ui.Context, p palette, block genui.Block) {
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		if block.Title != "" {
			ui.Text(c, block.Title).FontSize(fsUI).FontWeight(400).TextColor(p.text)
		}
		ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Stretch).Children(func() {
			a.canvasSide(c, p, block.Left)
			a.canvasSide(c, p, block.Right)
		})
		if block.Caption != "" {
			ui.Text(c, block.Caption).FontSize(fsCaption).TextColor(p.faint)
		}
	})
}

func (a *app) canvasSide(c *ui.Context, p palette, side *genui.Side) {
	ui.Column(c).Grow(1).MinWidth(0).Background(p.surface).BorderWidth(1).BorderColor(p.hairline).
		Radius(radiusCard).Padding(10, 12).Gap(6).Children(func() {
		ui.Text(c, side.Title).FontSize(fsUI).FontWeight(400).TextColor(p.accent)
		for _, item := range side.Items {
			ui.Row(c).FillWidth().Gap(7).AlignItems(ui.Start).Children(func() {
				ui.Text(c, "·").FontSize(fsRead).TextColor(p.accent)
				ui.Text(c, item).Grow(1).MinWidth(0).FontSize(fsUI).TextColor(p.body).LineHeight(1.55)
			})
		}
	})
}

func (a *app) canvasFlashcards(c *ui.Context, p palette, msg store.Message, blockIndex int, block genui.Block) {
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		if block.Title != "" {
			ui.Text(c, block.Title).FontSize(fsUI).FontWeight(400).TextColor(p.text)
		}
		for i := 0; i < len(block.Cards); i += 2 {
			ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Stretch).Children(func() {
				a.canvasFlashcard(c, p, msg, blockIndex, i, block.Cards[i])
				if i+1 < len(block.Cards) {
					a.canvasFlashcard(c, p, msg, blockIndex, i+1, block.Cards[i+1])
				} else {
					ui.Box(c).Grow(1).MinWidth(0)
				}
			})
		}
	})
}

func (a *app) canvasFlashcard(c *ui.Context, p palette, msg store.Message, blockIndex, cardIndex int, card genui.Card) {
	key := fmt.Sprintf("%s:%d:%d", msg.ID, blockIndex, cardIndex)
	flipped := a.flippedCards[key]
	bg, border := p.surface, p.hairline
	if flipped {
		bg, border = p.accentSoft, p.accentLine
	}
	button := ui.ButtonBase(c).Key(key).Grow(1).MinWidth(0).MinHeight(84).Padding(10, 12).
		Radius(radiusCard).Background(bg).BorderWidth(1).BorderColor(border).
		Label("记忆卡：" + card.Front).AlignItems(ui.Start)
	button.Children(func() {
		ui.Column(c).FillWidth().Gap(4).Children(func() {
			label := "点击翻面"
			if flipped {
				label = "答案"
			}
			ui.Text(c, label).FontSize(fsCaption).FontWeight(400).TextColor(p.accent)
			text := card.Front
			if flipped {
				text = card.Back
			}
			ui.Text(c, text).FontSize(fsRead).FontWeight(400).LineHeight(1.5).
				TextColor(p.text).MaxLines(4)
		})
	})
	if button.Hovered() {
		button.BorderColor(p.borderStrong)
	}
	if button.Clicked() {
		a.flippedCards[key] = !flipped
	}
}

func (a *app) canvasBars(c *ui.Context, p palette, block genui.Block) {
	maxValue := 0.0
	for _, value := range block.Values {
		maxValue = math.Max(maxValue, value)
	}
	if maxValue <= 0 {
		maxValue = 1
	}
	ui.Column(c).FillWidth().Gap(9).Children(func() {
		if block.Title != "" {
			ui.Text(c, block.Title).FontSize(fsUI).FontWeight(400).TextColor(p.text)
		}
		for i, label := range block.Labels {
			value := block.Values[i]
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				ui.Row(c).FillWidth().Children(func() {
					ui.Text(c, label).Grow(1).MinWidth(0).FontSize(fsUI).TextColor(p.body)
					ui.Text(c, formatBarValue(value)+block.Unit).FontSize(fsCaption).Font(monoFont).TextColor(p.muted)
				})
				percent := float32(value / maxValue * 100)
				if value > 0 && percent < 2 {
					percent = 2
				}
				ui.Row(c).FillWidth().Height(8).Radius(4).Background(p.surface2).Clip().Children(func() {
					ui.Box(c).WidthPercent(percent).Height(8).Radius(4).Background(p.accent)
				})
			})
		}
		if block.Caption != "" {
			ui.Text(c, block.Caption).FontSize(fsCaption).TextColor(p.faint)
		}
	})
}

func formatBarValue(v float64) string {
	if math.Abs(v-math.Round(v)) < 0.001 {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}
