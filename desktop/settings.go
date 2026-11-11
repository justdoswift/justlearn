package main

import (
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

func (a *app) settingsPage(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Overlay(c, func() {
		back := ui.Row(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Center()
		back.Background(ui.RGBA(0, 0, 0, 0.18*a.settingsShown))
		if a.settingsOpen {
			back.Modal()
		} else if a.settingsShown < 0.002 {
			back.PassThrough()
		}

		var panel ui.Element
		back.Children(func() {
			panel = ui.Column(c).Width(settingsWidth).MaxWidthPercent(92).Height(560).
				MaxHeightPercent(88).Radius(24).Background(p.panelBg).BorderWidth(1).
				BorderColor(p.hairline).Shadow(0, 18, 48, 0, p.shadow).Clip().
				Role(ui.RoleDialog).Label("设置")
			target := float32(0)
			if a.settingsOpen {
				target = 1
			}
			shown := target
			if !c.Preferences().ReduceMotion {
				shown = panel.AnimateTo("settings", target, panelSpring)
			}
			a.settingsShown = shown
			s := 0.965 + 0.035*shown
			panel.Scale(s, s).Opacity(shown).Translate(0, 10*(1-shown))
			panel.Children(func() {
				if a.settingsOpen || shown > 0.002 {
					a.settingsContent(c, t)
				}
			})
		})
		if a.settingsOpen && (panel.PressedOutside() || back.OverlayShortcut(0, ui.KeyEscape)) {
			a.closeSettings()
		}
		if !a.settingsOpen && a.settingsShown < 0.002 {
			back.Hidden(true)
		}
	})
}

func (a *app) settingsContent(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Row(c).FillWidth().FillHeight().Children(func() {
		ui.Column(c).Width(198).FillHeight().Padding(12).Gap(2).
			BorderWidth(0, 1, 0, 0).BorderColor(p.hairline).Children(func() {
			ui.Row(c).FillWidth().Height(38).PaddingX(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "设置").FontSize(13).FontWeight(400).TextColor(p.text)
			})
			a.settingsNavRow(c, t, 0, "外观", iconAuto)
			a.settingsNavRow(c, t, 1, "模型", iconConfig)
			a.settingsNavRow(c, t, 2, "关于", iconInfo)
			ui.Spacer(c)
			label := "JustLearn"
			if a.version != "" {
				label += " " + a.version
			}
			ui.Text(c, label).FontSize(fsCaption).TextColor(p.faint).Padding(8, 8)
		})

		ui.Column(c).Grow(1).FillHeight().Children(func() {
			ui.Row(c).FillWidth().Height(52).Padding(0, 14, 0, 20).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.settingsSectionTitle()).FontSize(fsTitle).FontWeight(400).TextColor(p.text)
				ui.Spacer(c)
				close := iconButton(c, p, iconClose, "关闭")
				if close.Clicked() {
					a.closeSettings()
				}
			})
			ui.Box(c).FillWidth().Height(1).Background(p.hairline)
			ui.Scroll(c).Fill().Children(func() {
				ui.Column(c).FillWidth().Padding(20, 22).Children(func() {
					ui.Column(c).FillWidth().MaxWidth(560).Children(func() {
						switch a.settingsSection {
						case 1:
							a.settingsModel(c, t)
						case 2:
							a.settingsAbout(c, t)
						default:
							a.settingsAppearance(c, t)
						}
					})
				})
			})
		})
	})
}

func (a *app) settingsSectionTitle() string {
	switch a.settingsSection {
	case 1:
		return "模型"
	case 2:
		return "关于"
	default:
		return "外观"
	}
}

func (a *app) settingsNavRow(c *ui.Context, t *ui.Theme, index int, label string, icon *ui.SVG) {
	p := paletteFor(t.Dark)
	selected := a.settingsSection == index
	row := ui.ButtonBase(c).FillWidth().Height(34).Radius(radiusControl).PaddingX(8).
		Transition(ui.ElementTransition{Colors: true, Duration: 140 * time.Millisecond})
	row.Children(func() {
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			color := p.muted
			if selected {
				color = p.accent
			}
			ui.Icon(c, icon).Size(14, 14).TextColor(color)
			ui.Text(c, label).FontSize(fsUI).FontWeight(400).TextColor(p.text)
		})
	})
	if selected {
		row.Background(p.navHover)
	} else if row.Hovered() {
		row.Background(p.navHover)
	}
	if row.Clicked() && !selected {
		a.settingsSection = index
		a.settingsSaved = false
		a.errText = ""
	}
}

func (a *app) settingsAppearance(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Row(c).FillWidth().Gap(20).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(16).Children(func() {
			ui.Text(c, "选择显示方式；跟随系统会随 macOS 自动切换。").
				FontSize(fsUI).TextColor(p.muted).LineHeight(1.6)
			ui.Column(c).FillWidth().Gap(8).Children(func() {
				a.appearanceRow(c, t, "system", "跟随系统", "自动匹配 macOS 外观", iconAuto)
				a.appearanceRow(c, t, "light", "浅色", "始终使用浅色外观", iconSun)
				a.appearanceRow(c, t, "dark", "深色", "始终使用深色外观", iconMoon)
			})
			ui.Text(c, "切换外观不会改变学习内容或对话记录。").
				FontSize(fsCaption).TextColor(p.muted).Padding(2, 0)
		})
		a.appearancePreview(c, t)
	})
}

func (a *app) appearancePreview(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).Width(200).Shrink(0).Height(286).Radius(14).Background(p.fieldBg).
		BorderWidth(1).BorderColor(p.hairline).Padding(12).Gap(8).Children(func() {
		ui.Text(c, "预览").FontSize(fsCaption).TextColor(p.faint)
		ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "✦").FontSize(9).TextColor(p.accent)
			ui.Text(c, "JustLearn").FontSize(fsCaption).FontWeight(400).TextColor(p.text)
		})
		ui.Text(c, "一张表到底长什么样").FontSize(14).Font(serifFont).FontWeight(400).TextColor(p.text).MaxLines(2)
		ui.Text(c, "数据库不是一堆文件，而是一组有约束的表格。").
			FontSize(fsCaption).TextColor(p.muted).LineHeight(1.55)
		ui.Text(c, "· 行是记录，列是字段").FontSize(fsCaption).TextColor(p.body)
		ui.Text(c, "· 主键唯一标识一行").FontSize(fsCaption).TextColor(p.body)
		ui.Spacer(c)
		ui.Row(c).FillWidth().Height(28).Radius(radiusControl).Background(p.surface2).
			BorderWidth(1).BorderColor(p.hairline).PaddingX(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "问点什么…").FontSize(fsCaption).TextColor(p.faint)
			ui.Spacer(c)
			ui.Box(c).Size(16, 16).Radius(8).Background(p.accent).Center().Children(func() {
				ui.Icon(c, iconArrowUp).Size(8, 8).TextColor(ui.Hex("#FFFFFF"))
			})
		})
	})
}

func (a *app) appearanceRow(c *ui.Context, t *ui.Theme, value, label, detail string, icon *ui.SVG) {
	p := paletteFor(t.Dark)
	selected := a.appearance == value
	bg := ui.Transparent
	border := p.hairline
	if selected {
		bg = p.fieldBg
		border = p.accentLine
	}
	row := ui.ButtonBase(c).FillWidth().Height(64).Radius(radiusCard).Padding(10, 12).
		Background(bg).BorderWidth(1).BorderColor(border).
		Transition(ui.ElementTransition{Colors: true, Duration: 180 * time.Millisecond})
	row.Children(func() {
		ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Center).Children(func() {
			iconBox := ui.Box(c).Size(32, 32).Radius(radiusControl).Background(p.fieldBg).Center()
			iconBox.Children(func() {
				ui.Icon(c, icon).Size(16, 16).TextColor(p.muted)
			})
			ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
				ui.Text(c, label).FontSize(fsUI).FontWeight(400).TextColor(p.text)
				ui.Text(c, detail).FontSize(fsCaption).TextColor(p.muted).MaxLines(1)
			})
			if selected {
				ui.Icon(c, iconCheck).Size(15, 15).TextColor(p.accent)
			}
		})
	})
	if row.Hovered() && !selected {
		row.Background(p.navHover)
	}
	if row.Clicked() {
		a.applyAppearance(value)
	}
}

func (a *app) settingsModel(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.Column(c).FillWidth().Gap(14).Children(func() {
		ui.Text(c, "填写任意 OpenAI 兼容接口。API Key 保存在系统钥匙串。").
			FontSize(fsUI).TextColor(p.muted).LineHeight(1.6)
		a.settingsField(c, t, "Base URL", &a.formBase, "https://opencode.ai/inference/go/openai/v1", false)
		a.settingsField(c, t, "模型", &a.formModel, "deepseek-v4.1-flash", false)
		keyHint := "sk-…"
		if cfgKey(a.store) != "" {
			keyHint = "已保存；留空则不修改"
		}
		a.settingsField(c, t, "API Key", &a.formKey, keyHint, true)
		if a.errText != "" {
			ui.Text(c, a.errText).FontSize(fsCaption).TextColor(p.rust).MaxLines(2).LineHeight(1.5)
		}
		ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Center).Padding(4, 0, 0, 0).Children(func() {
			save := ui.ButtonBase(c).Width(84).Height(30).Radius(9).
				Material(glass.Glass{Tint: p.accentStrong, Interactive: true}).Center()
			save.Children(func() {
				ui.Text(c, "保存").FontSize(fsUI).FontWeight(400).TextColor(ui.Hex("#FFFFFF"))
			})
			if save.Clicked() {
				a.saveSettings()
			}
			if a.settingsSaved {
				ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconCheck).Size(12, 12).TextColor(p.green)
					ui.Text(c, "已保存").FontSize(fsCaption).TextColor(p.green)
				})
			}
		})
	})
}

func (a *app) settingsField(c *ui.Context, t *ui.Theme, label string, value *string, placeholder string, password bool) {
	p := paletteFor(t.Dark)
	ui.Column(c).FillWidth().Gap(5).Children(func() {
		ui.Text(c, label).FontSize(fsCaption).TextColor(p.muted)
		input := ui.TextInput(c, value).FillWidth().Height(32).Radius(radiusControl).
			Background(p.fieldBg).BorderWidth(1).BorderColor(p.hairline).Placeholder(placeholder)
		if password {
			input.Password()
		}
	})
}

func (a *app) settingsAbout(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	version := a.version
	if version == "" {
		version = "开发版"
	}
	ui.Column(c).FillWidth().Gap(16).Children(func() {
		ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "✦").FontSize(24).TextColor(p.accent)
			ui.Column(c).Gap(2).Children(func() {
				ui.Text(c, "JustLearn").FontSize(fsSection).FontWeight(400).TextColor(p.text)
				ui.Text(c, version).FontSize(fsCaption).TextColor(p.faint)
			})
		})
		ui.Text(c, "原生桌面 AI 私人教师：在聊天里学、在聊天里测，整条学习路线汇成一张能力画像。").
			FontSize(fsUI).TextColor(p.muted).LineHeight(1.65)
		ui.Box(c).FillWidth().Height(1).Background(p.hairline)
		a.aboutRow(c, t, "本地数据", "SQLite，保存在这台 Mac 上。")
		a.aboutRow(c, t, "模型", "通过任意 OpenAI 兼容接口调用。")
		a.aboutRow(c, t, "更新", "JustLearn → 检查更新…")
		a.aboutRow(c, t, "协议", "MIT")
	})
}

func (a *app) aboutRow(c *ui.Context, t *ui.Theme, label, detail string) {
	p := paletteFor(t.Dark)
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Text(c, label).Width(72).FontSize(fsUI).TextColor(p.text)
		ui.Text(c, detail).Grow(1).FontSize(fsUI).TextColor(p.muted).LineHeight(1.5)
	})
}
