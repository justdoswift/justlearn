package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

type lessonPayload struct {
	NodeID string `json:"nodeId"`
	Kicker string `json:"kicker"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Recall struct {
		Prompt string `json:"prompt"`
		Answer string `json:"answer"`
	} `json:"recall"`
	Code      string   `json:"code"`
	Mnemonic  string   `json:"mnemonic"`
	KeyPoints []string `json:"keyPoints"`
}

type quizPayload struct {
	NodeID    string `json:"nodeId"`
	Questions []struct {
		ID          string   `json:"id"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		AnswerIndex int      `json:"answerIndex"`
		Explanation string   `json:"explanation"`
	} `json:"questions"`
}

func (a *app) mdOptions(t *ui.Theme) *markdown.Options {
	p := paletteFor(t.Dark)
	theme := markdown.DefaultTheme(t)
	theme.FontSize = fsRead
	theme.LineHeight = 1.72
	theme.Font = uiFont
	theme.Text = p.body
	theme.Muted = p.muted
	theme.Link = p.blue
	theme.Strong = p.text
	theme.Code = p.text
	theme.CodeFont = monoFont
	theme.CodeBackground = p.fieldBg
	theme.CodeBlockBackground = p.surface2
	theme.CodeBlockBorder = p.surface2
	theme.CodeBlockRadius = 4
	theme.BlockSpacing = 0.6
	a.md.Theme = theme
	return a.md
}

func (a *app) main(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).Fill().FillHeight().Background(p.canvas).Children(func() {
		a.chat(c, t)
		a.composerBox(c, t)
	})
}

func (a *app) chat(c *ui.Context, t *ui.Theme) {
	if len(a.messages) == 0 && !a.streaming {
		ui.Column(c).Fill().Center().Children(func() {
			a.emptyChat(c, t)
		})
		return
	}
	rows := len(a.messages)
	if a.streaming {
		rows++
	}
	ui.List(c, &a.chatList, rows, func(i int) {
		a.chatRow(c, t, i)
	}).Fill()
}

func (a *app) chatRow(c *ui.Context, t *ui.Theme, i int) {
	ui.Column(c).FillWidth().Padding(18, 7).AlignItems(ui.Center).Children(func() {
		ui.Column(c).FillWidth().MaxWidth(680).Children(func() {
			if len(a.messages) == 0 {
				if a.streaming {
					a.streamingRow(c, t)
				} else {
					a.emptyChat(c, t)
				}
				return
			}
			if i >= len(a.messages) {
				a.streamingRow(c, t)
				return
			}
			msg := a.messages[i]
			switch {
			case msg.Role == "user":
				a.userBubble(c, t, msg.Text)
			case msg.Kind == "lesson":
				a.lessonSection(c, t, msg)
			case msg.Kind == "quiz":
				a.quizSection(c, t, msg)
			default:
				markdown.View(c, msg.Text, a.mdOptions(t)).FillWidth()
			}
		})
	})
}

func (a *app) emptyChat(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).FillWidth().AlignItems(ui.Center).Gap(8).Children(func() {
		ui.Text(c, "✦").FontSize(16).TextColor(p.blue)
		ui.Text(c, "开始一段新的学习").FontSize(fsSection).FontWeight(600).TextColor(p.text)
		ui.Text(c, "告诉教练你想学什么，路线和热力图会在聊天里长出来。").
			FontSize(fsUI).TextColor(p.muted).TextAlign(ui.Center).MaxWidth(420).LineHeight(1.5)
		ui.Row(c).Gap(6).PaddingY(4).Children(func() {
			for _, suggestion := range []string{"转行数据分析", "通过日语 N2", "读懂财务报表"} {
				text := suggestion
				chip := ui.ButtonBase(c).Padding(4, 8).Radius(radiusBadge).Children(func() {
					ui.Text(c, text).FontSize(fsCaption).TextColor(p.muted)
				})
				if chip.Hovered() {
					chip.Background(p.navHover)
				}
				if chip.Clicked() && !a.streaming {
					a.send(text)
				}
			}
		})
	})
}

func (a *app) userBubble(c *ui.Context, t *ui.Theme, text string) {
	p := paletteFor(t.Dark)
	ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
		ui.Box(c).MaxWidth(480).Padding(6, 10).Radius(radiusControl).Background(p.navActive).Children(func() {
			ui.Text(c, text).FontSize(fsRead).LineHeight(1.5).TextColor(p.text)
		})
	})
}

func (a *app) streamingRow(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	if strings.TrimSpace(a.streamText) == "" {
		ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
			ui.Spinner(c)
			ui.Text(c, "正在想…").FontSize(fsUI).TextColor(p.muted)
		})
		return
	}
	markdown.Stream(c, a.streamText, a.mdOptions(t)).FillWidth()
}

func (a *app) kicker(c *ui.Context, t *ui.Theme, text string) {
	p := paletteFor(t.Dark)
	ui.Text(c, text).FontSize(fsCaption).TextColor(p.faint).MaxLines(1)
}

func (a *app) lessonSection(c *ui.Context, t *ui.Theme, msg store.Message) {
	p := paletteFor(t.Dark)
	var payload lessonPayload
	_ = json.Unmarshal([]byte(msg.Payload), &payload)
	ui.Column(c).FillWidth().BorderWidth(0, 0, 0, 2).BorderColor(p.blue).
		Padding(8, 0, 2, 14).Gap(8).Children(func() {
		a.kicker(c, t, payload.Kicker)
		if payload.Title != "" {
			ui.Text(c, payload.Title).FontSize(fsTitle).FontWeight(600).LineHeight(1.35).TextColor(p.text)
		}
		markdown.View(c, payload.Body, a.mdOptions(t)).FillWidth()
		if payload.Code != "" {
			ui.Box(c).FillWidth().Background(p.fieldBg).Radius(radiusBadge).Padding(8, 10).Children(func() {
				ui.Text(c, payload.Code).FontSize(fsCaption).Font(monoFont).TextColor(p.body)
			})
		}
		if payload.Recall.Prompt != "" {
			ui.Box(c).FillWidth().Background(p.recallBg).Radius(radiusBadge).
				BorderWidth(0, 0, 0, 2).BorderColor(p.blue).Padding(8, 10).Children(func() {
				ui.Column(c).FillWidth().Gap(3).Children(func() {
					ui.Text(c, "主动回忆").FontSize(fsCaption).FontWeight(600).TextColor(p.blue)
					ui.Text(c, payload.Recall.Prompt).FontSize(fsRead).FontWeight(600).TextColor(p.text)
					ui.Text(c, "参考答案："+payload.Recall.Answer).FontSize(fsUI).TextColor(p.muted)
				})
			})
		}
		if payload.Mnemonic != "" {
			ui.Text(c, "记："+payload.Mnemonic).FontSize(fsUI).TextColor(p.muted)
		}
		for _, point := range payload.KeyPoints {
			ui.Row(c).FillWidth().Gap(6).Children(func() {
				ui.Text(c, "·").FontSize(fsUI).Bold().TextColor(p.blue)
				ui.Text(c, point).FontSize(fsUI).TextColor(p.body).LineHeight(1.5)
			})
		}
	})
}

func (a *app) quizSection(c *ui.Context, t *ui.Theme, msg store.Message) {
	p := paletteFor(t.Dark)
	var payload quizPayload
	_ = json.Unmarshal([]byte(msg.Payload), &payload)
	ui.Column(c).FillWidth().BorderWidth(0, 0, 0, 2).BorderColor(p.blue).
		Padding(8, 0, 2, 14).Gap(8).Children(func() {
		a.kicker(c, t, "小测")
		for qi, q := range payload.Questions {
			chosen, answered := a.quizChoices[q.ID]
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				ui.Textf(c, "%d. %s", qi+1, q.Prompt).FontSize(fsTitle).FontWeight(600).
					LineHeight(1.45).TextColor(p.text)
				for oi, option := range q.Options {
					letter := string(rune('A' + oi))
					correct := answered && oi == q.AnswerIndex
					wrong := answered && chosen == oi && oi != q.AnswerIndex
					fg := p.body
					bg := ui.Transparent
					if correct {
						fg, bg = p.blue, p.blueSoft
					} else if wrong {
						fg, bg = p.rust, p.rustSoft
					} else if answered {
						fg = p.faint
					}
					opt := ui.ButtonBase(c.Key(q.ID+":opt:"+strconv.Itoa(oi))).FillWidth().
						Padding(3, 6).Radius(radiusBadge).Background(bg).Children(func() {
						ui.Row(c).FillWidth().Gap(7).AlignItems(ui.Center).Children(func() {
							ui.Text(c, letter).FontSize(fsCaption).Font(monoFont).TextColor(fg)
							ui.Text(c, option).Grow(1).FontSize(fsRead).TextColor(fg).MaxLines(2)
							if correct {
								ui.Text(c, "正确答案").FontSize(fsCaption).TextColor(p.blue)
							} else if wrong {
								ui.Text(c, "你的选择").FontSize(fsCaption).TextColor(p.rust)
							}
						})
					})
					if !answered && !a.streaming && opt.Hovered() {
						opt.Background(p.fieldBg)
					}
					if answered {
						opt.Disabled(true)
					}
					if !answered && !a.streaming && opt.Clicked() {
						a.quizChoices[q.ID] = oi
						a.send(fmt.Sprintf("第 %d 题（%s）我选 %s：%s", qi+1, q.Prompt, letter, option))
					}
				}
				if answered {
					verdict := "答对了。"
					color := p.blue
					if chosen != q.AnswerIndex {
						verdict = "不对，正确答案是 " + string(rune('A'+q.AnswerIndex)) + "。"
						color = p.rust
					}
					ui.Text(c, verdict+q.Explanation).FontSize(fsCaption).TextColor(color).LineHeight(1.5)
				}
			})
		}
	})
}

func (a *app) composerBox(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	if c.Shortcut(ui.Cmd, ui.KeyEnter) {
		a.send(a.draft)
	}
	top := p.hairline
	if a.streaming {
		top = p.blue
	}
	ui.Column(c).FillWidth().AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).
		BorderColor(top).Background(p.surface2).Padding(6, 14, 8, 14).Children(func() {
		ui.Row(c).FillWidth().MaxWidth(760).Gap(8).AlignItems(ui.Center).Children(func() {
			plus := ui.ButtonBase(c).Size(26, 26).Radius(radiusBadge).Label("导入文本").
				Tooltip("导入文本").Center().Children(func() {
				ui.Text(c, "+").FontSize(15).TextColor(p.muted)
			})
			if plus.Hovered() {
				plus.Background(p.fieldBg)
			}
			if plus.Clicked() {
				a.attachText()
			}

			placeholder := "回答，或问点什么…"
			if len(a.messages) == 0 {
				placeholder = "问教练任何问题…"
			}
			field := ui.TextAreaBase(c, &a.draft).Grow(1).Lines(1, 5).FontSize(fsRead).
				Placeholder(placeholder).Bind(&a.composer)
			if a.focusDraft {
				field.Focus()
				a.focusDraft = false
			}

			if a.streaming {
				stop := ui.ButtonBase(c).Size(28, 28).Radius(14).Background(p.blue).
					Label("停止").Tooltip("停止").Center().Children(func() {
					ui.Text(c, "■").FontSize(9).TextColor(ui.Hex("#FFFFFF"))
				})
				if stop.Clicked() {
					a.cancelSend()
				}
			} else {
				disabled := strings.TrimSpace(a.draft) == ""
				bg, fg, border := p.blue, ui.Hex("#FFFFFF"), p.blue
				if disabled {
					bg, fg, border = p.fieldBg, p.faint, p.hairline
				}
				send := ui.ButtonBase(c).Size(28, 28).Radius(14).Background(bg).
					BorderWidth(1).BorderColor(border).Label("发送").Tooltip("发送").Center().Children(func() {
					ui.Text(c, "↑").FontSize(14).FontWeight(600).TextColor(fg)
				})
				if disabled {
					send.IgnoreInput(true)
				} else if send.Hovered() {
					send.Background(withAlpha(p.blue, 0.9))
				}
				if !disabled && send.Clicked() {
					a.send(a.draft)
				}
			}
		})
		if a.errText != "" {
			ui.Column(c).FillWidth().MaxWidth(760).Padding(3, 0, 0, 0).Children(func() {
				ui.Text(c, a.errText).FontSize(fsCaption).TextColor(p.rust).MaxLines(2)
			})
		}
	})
}

func (a *app) modelPanel(c *ui.Context, t *ui.Theme, p palette, panel ui.Element) {
	panel.Background(p.surface).Radius(radiusCard).BorderWidth(1).BorderColor(p.hairline).
		Padding(4).Width(230).Shadow(0, 12, 30, 0, withAlpha(ui.Hex("#000000"), 0.16))
	panel.Children(func() {
		ui.Column(c).FillWidth().Gap(1).Children(func() {
			if a.modelsLoading {
				ui.Row(c).Padding(6, 6).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Spinner(c)
					ui.Text(c, "读取模型…").FontSize(fsCaption).TextColor(p.faint)
				})
			}
			current := cfgModel(a.store)
			list := make([]string, 0, len(a.models)+1)
			seen := map[string]bool{}
			list = append(list, current)
			seen[current] = true
			for _, id := range a.models {
				if id != "" && !seen[id] && len(list) < 20 {
					list = append(list, id)
					seen[id] = true
				}
			}
			for _, id := range list {
				model := id
				row := ui.ButtonBase(c).FillWidth().Height(24).Radius(radiusBadge).PaddingX(6).
					Children(func() {
						ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
							ui.Text(c, prettyModel(model)).FontSize(fsCaption).TextColor(p.text).Grow(1).MaxLines(1)
							if model == current {
								ui.Text(c, "✓").FontSize(fsCaption).TextColor(p.blue)
							}
						})
					})
				if row.Hovered() {
					row.Background(p.fieldBg)
				}
				if row.Clicked() {
					a.selectModel(model)
				}
			}
			hairline(c, p.hairline)
			settings := ui.ButtonBase(c).FillWidth().Height(24).Radius(radiusBadge).PaddingX(6).
				Justify(ui.Start).Children(func() {
				ui.Text(c, "在设置中配置…").FontSize(fsCaption).TextColor(p.muted)
			})
			if settings.Hovered() {
				settings.Background(p.fieldBg)
			}
			if settings.Clicked() {
				a.modelMenuOpen = false
				a.openSettings()
			}
		})
	})
}
