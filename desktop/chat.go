package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/plugins/glass"
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
	theme.LineHeight = 1.7
	theme.Font = uiFont
	theme.Text = p.body
	theme.Muted = p.muted
	theme.Link = p.accent
	theme.LinkWeight = 400
	theme.Strong = p.text
	theme.StrongWeight = 400
	for i := range theme.Headings {
		theme.Headings[i].Weight = 400
	}
	theme.Code = p.text
	theme.CodeFont = monoFont
	theme.CodeBackground = p.fieldBg
	theme.CodeBlockBackground = p.surface2
	theme.CodeBlockBorder = p.hairline
	theme.CodeBlockRadius = radiusControl
	theme.CodeSize = 0.9
	theme.CodeLineHeight = 1.6
	theme.CodeBlockPadding = 0.85
	theme.BlockSpacing = 0.95
	theme.ItemSpacing = 0.4
	theme.ListIndent = 1.5
	theme.Rule = p.hairline
	theme.QuoteBar = p.accent
	theme.QuoteText = p.body
	theme.QuoteBarWidth = 3
	a.md.Theme = theme
	return a.md
}

func (a *app) main(c *ui.Context, t *ui.Theme) {
	if c.Shortcut(ui.Cmd, ui.KeyEnter) {
		a.send(a.draft)
	}
	if a.home {
		a.homeView(c, t)
		return
	}
	a.chatView(c, t)
}

const (
	chatTopInset    float32 = 16
	chatBottomInset float32 = 168
)

func (a *app) homeView(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).FillWidth().FillHeight().Justify(ui.Center).AlignItems(ui.Center).
		Padding(0, 24, 48, 24).Children(func() {
		ui.Column(c).Width(readingWidth).AlignItems(ui.Center).Gap(12).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "✦").FontSize(13).TextColor(p.accent)
				ui.Text(c, "JustLearn").FontSize(13).FontWeight(400).TextColor(p.text)
			})
			a.composerShell(c, t)
			if len(a.messages) > 0 && !a.streaming {
				more := ui.ButtonBase(c).Height(28).PaddingX(8).Radius(radiusControl).
					Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
				more.Children(func() {
					ui.Text(c, "继续上次 →").FontSize(fsUI).TextColor(p.muted)
				})
				if more.Hovered() {
					more.Background(p.navHover)
				}
				if more.Clicked() {
					a.continueChat()
				}
			}
		})
	})
}

func (a *app) chatView(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Box(c).Fill().Children(func() {
		ui.Column(c).Fill().Padding(0, 0, 0, 0).Children(func() {
			a.chat(c, t)
		})
		ui.Box(c).Absolute().Left(0).Right(0).Bottom(0).Height(chatBottomInset).PassThrough().
			Material(glass.ScrollEdge{Bottom: true, Background: p.backdropA})
		ui.Column(c).Absolute().Left(0).Right(0).Bottom(0).AlignItems(ui.Center).
			Padding(0, 24, 12, 24).Children(func() {
			ui.Column(c).FillWidth().MaxWidth(readingWidth).Gap(8).Children(func() {
				var modelButton ui.Element
				ui.Row(c).FillWidth().Justify(ui.End).PaddingX(4).Children(func() {
					modelButton = ui.ButtonBase(c).Height(24).Radius(radiusControl).PaddingX(5).
						Tooltip("选择模型").
						Transition(ui.ElementTransition{Colors: true, Duration: 140 * time.Millisecond})
					modelButton.Children(func() {
						ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
							ui.Text(c, prettyModel(cfgModel(a.store))).FontSize(fsUI).
								TextColor(p.muted).MaxLines(1).MaxWidth(240)
							ui.Icon(c, iconChevron).Size(10, 10).TextColor(p.faint)
						})
					})
					if modelButton.Hovered() {
						modelButton.Background(p.navHover)
					}
					if modelButton.Clicked() {
						a.modelMenuOpen = !a.modelMenuOpen
						if a.modelMenuOpen {
							a.loadModels()
						}
					}
				})
				a.composerShell(c, t)
				if a.errText != "" {
					ui.Text(c, a.errText).FontSize(fsCaption).TextColor(p.rust).MaxLines(2).Padding(0, 2)
				}
				ui.PopoverBase(c, modelButton, &a.modelMenuOpen, func(panel ui.Element) {
					a.modelPanel(c, t, p, panel, modelButton)
				})
			})
		})
	})
}

func (a *app) chat(c *ui.Context, t *ui.Theme) {
	if a.chatScrollToEnd {
		a.chatList.ScrollToEnd()
		a.chatScrollToEnd = false
	}
	rows := 1 + len(a.messages) + 1
	if a.streaming {
		rows++
	}
	ui.List(c, &a.chatList, rows, func(i int) {
		switch {
		case i == 0:
			ui.Box(c).FillWidth().Height(chatTopInset)
		case i == rows-1:
			ui.Box(c).FillWidth().Height(chatBottomInset)
		default:
			idx := i - 1
			if a.streaming && idx == len(a.messages) {
				ui.Column(c).FillWidth().Padding(12, 24, 9, 24).AlignItems(ui.Center).Children(func() {
					ui.Column(c).FillWidth().MaxWidth(readingWidth).Children(func() {
						a.streamingRow(c, t)
					})
				})
				return
			}
			a.chatRow(c, t, idx)
		}
	}).FillWidth().FillHeight().ScrollbarInsets(chatTopInset, 0, chatBottomInset, 0)
}

func (a *app) chatRow(c *ui.Context, t *ui.Theme, i int) {
	ui.Column(c).FillWidth().Padding(12, 24, 9, 24).AlignItems(ui.Center).Children(func() {
		ui.Column(c).FillWidth().MaxWidth(readingWidth).Children(func() {
			if i < 0 || i >= len(a.messages) {
				return
			}
			msg := a.messages[i]
			switch {
			case msg.Role == "user":
				a.userMessage(c, t, msg.Text)
			case msg.Kind == "lesson":
				a.lessonSection(c, t, msg)
			case msg.Kind == "quiz":
				a.quizSection(c, t, msg)
			case msg.Kind == "ui":
				a.canvasSection(c, t, msg)
			default:
				a.textMessage(c, t, msg)
			}
		})
	})
}

func (a *app) userMessage(c *ui.Context, t *ui.Theme, text string) {
	p := paletteFor(t.Dark)
	ui.Row(c).FillWidth().Children(func() {
		ui.Box(c).MaxWidth(560).BorderWidth(0, 0, 0, 3).BorderColor(p.accent).
			Padding(2, 0, 2, 14).Children(func() {
			ui.Text(c, text).FontSize(fsRead).LineHeight(1.6).TextColor(p.body)
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

// textMessage keeps ordinary assistant prose as Markdown, but gives older
// plain-text choice lists the same row geometry as today's quiz.
func (a *app) textMessage(c *ui.Context, t *ui.Theme, msg store.Message) {
	if _, _, _, ok := splitLegacyOptions(msg.Text); !ok {
		markdown.View(c, msg.Text, a.mdOptions(t)).FillWidth()
		return
	}
	p := paletteFor(t.Dark)
	ui.Column(c).FillWidth().Gap(10).Children(func() {
		rest := msg.Text
		for {
			before, options, after, ok := splitLegacyOptions(rest)
			if !ok {
				if strings.TrimSpace(rest) != "" {
					markdown.View(c, rest, a.mdOptions(t)).FillWidth()
				}
				return
			}
			if strings.TrimSpace(before) != "" {
				markdown.View(c, before, a.mdOptions(t)).FillWidth()
			}
			ui.Column(c).FillWidth().Gap(2).Children(func() {
				for _, option := range options {
					ui.Box(c).FillWidth().Padding(6, 8).Radius(radiusControl).Children(func() {
						optionContent(c, p, option.Letter, option.Text, option.Code, p.body, p.muted, "")
					})
				}
			})
			if strings.TrimSpace(after) == "" {
				return
			}
			rest = after
		}
	})
}

func (a *app) kicker(c *ui.Context, t *ui.Theme, text string) {
	p := paletteFor(t.Dark)
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(5, 5).Radius(2.5).Background(p.accent)
		ui.Text(c, text).FontSize(fsCaption).Font(monoFont).TextColor(p.muted).MaxLines(1)
	})
}

func (a *app) lessonSection(c *ui.Context, t *ui.Theme, msg store.Message) {
	p := paletteFor(t.Dark)
	var payload lessonPayload
	_ = json.Unmarshal([]byte(msg.Payload), &payload)
	ui.Column(c).FillWidth().Padding(0, 0, 4, 0).Gap(12).Children(func() {
		a.kicker(c, t, payload.Kicker)
		if payload.Title != "" {
			ui.Text(c, payload.Title).FontSize(21).Font(serifFont).FontWeight(400).LineHeight(1.35).TextColor(p.text)
		}
		markdown.View(c, payload.Body, a.mdOptions(t)).FillWidth()
		if payload.Code != "" {
			a.codeBlock(c, p, msg.ID+":lesson-code", "", payload.Code, "")
		}
		if payload.Recall.Prompt != "" {
			ui.Box(c).FillWidth().Background(p.recallBg).Radius(radiusControl).
				BorderWidth(0, 0, 0, 2).BorderColor(p.accentLine).Padding(9, 11).Children(func() {
				ui.Column(c).FillWidth().Gap(3).Children(func() {
					ui.Text(c, "主动回忆").FontSize(fsCaption).FontWeight(400).TextColor(p.accent)
					ui.Text(c, payload.Recall.Prompt).FontSize(fsRead).FontWeight(400).TextColor(p.text)
					ui.Text(c, "参考答案："+payload.Recall.Answer).FontSize(fsUI).TextColor(p.muted)
				})
			})
		}
		if payload.Mnemonic != "" {
			ui.Text(c, "记："+payload.Mnemonic).FontSize(fsUI).TextColor(p.muted)
		}
		for _, point := range payload.KeyPoints {
			ui.Row(c).FillWidth().Gap(7).Children(func() {
				ui.Text(c, "·").FontSize(fsRead).TextColor(p.faint)
				ui.Text(c, point).FontSize(fsRead).TextColor(p.body).LineHeight(1.55)
			})
		}
	})
}

func (a *app) quizSection(c *ui.Context, t *ui.Theme, msg store.Message) {
	p := paletteFor(t.Dark)
	var payload quizPayload
	_ = json.Unmarshal([]byte(msg.Payload), &payload)
	ui.Column(c).FillWidth().Padding(0, 0, 4, 0).Gap(11).Children(func() {
		a.kicker(c, t, "小测")
		for qi, q := range payload.Questions {
			chosen, answered := a.quizChoices[q.ID]
			ui.Column(c).FillWidth().Gap(6).Children(func() {
				ui.Textf(c, "%d. %s", qi+1, q.Prompt).FontSize(16).FontWeight(400).
					LineHeight(1.55).TextColor(p.text)
				for oi, option := range q.Options {
					letter := string(rune('A' + oi))
					correct := answered && oi == q.AnswerIndex
					wrong := answered && chosen == oi && oi != q.AnswerIndex
					fg := p.body
					bg := ui.Transparent
					codeOption := looksLikeCode(option)
					if correct {
						fg, bg = p.green, p.greenSoft
					} else if wrong {
						fg, bg = p.rust, p.rustSoft
					} else if answered {
						fg = p.faint
					}
					trailing, trailingColor := "", p.muted
					if correct {
						trailing, trailingColor = "正确答案", p.green
					} else if wrong {
						trailing, trailingColor = "你的选择", p.rust
					}
					opt := ui.ButtonBase(c.Key(q.ID+":opt:"+strconv.Itoa(oi))).FillWidth().
						Padding(6, 8).Radius(radiusControl).Background(bg)
					opt.Children(func() {
						optionContent(c, p, letter, option, codeOption, fg, trailingColor, trailing)
					})
					if !answered && !a.streaming && opt.Hovered() {
						opt.Background(p.navHover)
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
					color := p.green
					if chosen != q.AnswerIndex {
						verdict = "不对，正确答案是 " + string(rune('A'+q.AnswerIndex)) + "。"
						color = p.rust
					}
					ui.Text(c, verdict+q.Explanation).FontSize(fsCaption).TextColor(color).LineHeight(1.55)
				}
			})
		}
	})
}

type composerMatch struct{}

// composerShell is the same glass field in both states. The shared Match
// identity lets it fly from the middle of the window to the bottom when
// the first message starts the conversation.
func (a *app) composerShell(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	panel := ui.Column(c).FillWidth().Radius(24).Material(composerGlass(t.Dark)).
		BorderWidth(1).BorderColor(withAlpha(p.borderStrong, 0.55)).
		Shadow(0, 6, 16, 0, withAlpha(ui.Hex("#000000"), 0.08)).
		Padding(6, 10).
		Label("输入区").Transition(ui.ElementTransition{Animation: &ui.Snappy, Match: composerMatch{}})
	panel.Children(func() {
		placeholder := "回答，或继续追问…"
		if a.home {
			placeholder = "今天想学点什么？"
		}
		ui.Row(c).FillWidth().MinHeight(34).AlignItems(ui.Center).Gap(6).Children(func() {
			plus := ui.ButtonBase(c).Size(26, 26).Radius(13).Label("导入文本").
				Tooltip("导入文本").Center().
				Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
			plus.Children(func() {
				ui.Icon(c, iconPlus).Size(14, 14).TextColor(p.muted)
			})
			if plus.Hovered() {
				plus.Background(p.navHover)
			}
			if plus.Clicked() {
				a.attachText()
			}

			input := ui.TextAreaBase(c, &a.draft).Grow(1).MinWidth(0).Lines(1, 6).FontSize(fsRead).
				Placeholder(placeholder).Bind(&a.composer)
			if a.focusDraft {
				input.Focus()
				a.focusDraft = false
			}
			if input.Focused() {
				panel.BorderColor(withAlpha(p.accent, 0.55))
			}

			if a.streaming {
				stop := ui.ButtonBase(c).Size(26, 26).Radius(13).
					Material(glass.Glass{Tint: p.accentStrong, Interactive: true}).
					Label("停止").Tooltip("停止").Center()
				stop.Children(func() {
					ui.Icon(c, iconStop).Size(9, 9).TextColor(ui.Hex("#FFFFFF"))
				})
				pose := ui.Pose{}
				if stop.Hovered() {
					pose.Scale = 1.05
				}
				stop.AnimatePose(pose, ui.Snappy)
				if stop.Clicked() {
					a.cancelSend()
				}
			} else if strings.TrimSpace(a.draft) == "" {
				// Keep the slot so the field never changes width when the
				// send button appears; leave it out of the accessibility
				// tree while it has nothing to do.
				ui.Box(c).Size(26, 26)
			} else {
				send := ui.ButtonBase(c).Size(26, 26).Radius(13).
					Material(glass.Glass{Tint: p.accentStrong, Interactive: true}).
					Label("发送").Tooltip("发送").Center().
					Transition(ui.ElementTransition{
						Animation: &ui.Snappy,
						Enter:     &ui.Motion{Scale: 0.84},
						Exit:      &ui.Motion{Scale: 0.84},
					})
				send.Children(func() {
					ui.Icon(c, iconArrowUp).Size(12, 12).TextColor(ui.Hex("#FFFFFF"))
				})
				pose := ui.Pose{}
				if send.Hovered() {
					pose.Scale = 1.05
				}
				send.AnimatePose(pose, ui.Snappy)
				if send.Clicked() {
					a.send(a.draft)
				}
			}
		})
	})
}

func (a *app) modelPanel(c *ui.Context, t *ui.Theme, p palette, panel, anchor ui.Element) {
	panel.AttachTo(anchor, ui.AnchorTopEnd, ui.AnchorBottomEnd).Margin(0, 0, 6, 0).
		Background(p.panelBg).BorderWidth(1).BorderColor(p.hairline).Radius(16).Padding(6).Width(250).
		Shadow(0, 18, 46, 0, withAlpha(ui.Hex("#000000"), 0.28)).
		Transition(ui.ElementTransition{
			Animation: &ui.Snappy,
			Enter:     &ui.Motion{Scale: 0.97, Y: 4},
			Exit:      &ui.Motion{Scale: 0.97, Y: 4},
		})
	panel.Children(func() {
		ui.Column(c).FillWidth().Gap(1).Children(func() {
			ui.Text(c, "模型").FontSize(fsCaption).TextColor(p.muted).Padding(5, 8)
			if a.modelsLoading {
				ui.Row(c).Padding(6, 8).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Spinner(c)
					ui.Text(c, "读取模型…").FontSize(fsCaption).TextColor(p.muted)
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
				row := ui.ButtonBase(c).FillWidth().Height(30).Radius(radiusControl).PaddingX(8).
					Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
				row.Children(func() {
					ui.Row(c).FillWidth().Gap(6).AlignItems(ui.Center).Children(func() {
						ui.Text(c, prettyModel(model)).FontSize(fsUI).TextColor(p.text).Grow(1).MaxLines(1)
						if model == current {
							ui.Icon(c, iconCheck).Size(12, 12).TextColor(p.accent)
						}
					})
				})
				if row.Hovered() {
					row.Background(p.navHover)
				}
				if row.Clicked() {
					a.selectModel(model)
				}
			}
			ui.Box(c).FillWidth().Height(1).Background(p.hairline).Margin(5, 3)
			settings := ui.ButtonBase(c).FillWidth().Height(30).Radius(radiusControl).PaddingX(8).
				Justify(ui.Start).Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
			settings.Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconConfig).Size(13, 13).TextColor(p.muted)
					ui.Text(c, "模型设置…").FontSize(fsUI).TextColor(p.muted)
				})
			})
			if settings.Hovered() {
				settings.Background(p.navHover)
			}
			if settings.Clicked() {
				a.modelMenuOpen = false
				a.openSettingsAt(1)
			}
		})
	})
}
