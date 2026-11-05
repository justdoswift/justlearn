package main

import "github.com/egoist/mygo/ui"

func (a *app) settingsDialog(c *ui.Context, t *ui.Theme) {
	p := paletteFor(t.Dark)
	ui.DialogBase(c, &a.settingsOpen, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.35))
		panel.Background(p.surface).Radius(radiusCard).BorderWidth(1).BorderColor(p.hairline).
			Shadow(0, 18, 48, 0, withAlpha(ui.Hex("#000000"), 0.22)).
			Padding(16).Width(380).Children(func() {
			ui.Column(c).FillWidth().Gap(8).Children(func() {
				ui.Text(c, "模型设置").FontSize(fsTitle).FontWeight(600).TextColor(p.text)
				ui.Text(c, "OpenAI 兼容接口。API Key 保存在系统钥匙串。").
					FontSize(fsCaption).TextColor(p.muted).LineHeight(1.5)
				ui.Column(c).FillWidth().Gap(3).Children(func() {
					ui.Text(c, "Base URL").FontSize(fsCaption).TextColor(p.muted)
					ui.TextInput(c, &a.formBase).FillWidth().
						Placeholder("https://opencode.ai/inference/go/openai/v1")
				})
				ui.Column(c).FillWidth().Gap(3).Children(func() {
					ui.Text(c, "模型").FontSize(fsCaption).TextColor(p.muted)
					ui.TextInput(c, &a.formModel).FillWidth().
						Placeholder("deepseek-v4.1-flash")
				})
				keyHint := "sk-…"
				if cfgKey(a.store) != "" {
					keyHint = "已从环境变量读取，留空则不修改"
				}
				ui.Column(c).FillWidth().Gap(3).Children(func() {
					ui.Text(c, "API Key").FontSize(fsCaption).TextColor(p.muted)
					ui.TextInput(c, &a.formKey).FillWidth().Password().Placeholder(keyHint)
				})
				if a.errText != "" {
					ui.Text(c, a.errText).FontSize(fsCaption).TextColor(p.rust).MaxLines(2)
				}
				ui.Row(c).FillWidth().Gap(6).Justify(ui.End).PaddingY(2).Children(func() {
					cancel := ui.ButtonBase(c).Width(60).Height(24).Radius(radiusBadge).BorderWidth(1).
						BorderColor(p.hairline).Center().Children(func() {
						ui.Text(c, "取消").FontSize(fsCaption).TextColor(p.muted)
					})
					if cancel.Hovered() {
						cancel.BorderColor(p.borderStrong)
					}
					if cancel.Clicked() {
						a.settingsOpen = false
						a.errText = ""
					}
					save := ui.ButtonBase(c).Width(60).Height(24).Radius(radiusBadge).Background(p.blueStrong).Center().
						Children(func() {
							ui.Text(c, "保存").FontSize(fsCaption).FontWeight(600).TextColor(ui.Hex("#FFFFFF"))
						})
					if save.Hovered() {
						save.Background(withAlpha(p.blueStrong, 0.9))
					}
					if save.Clicked() {
						a.saveSettings()
					}
				})
			})
		})
	})
}
