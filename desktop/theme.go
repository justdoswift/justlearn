package main

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// Type scale: meta 11 / UI 12 / reading 14 / title 17 / section 20 / display 24.
const (
	fsCaption float32 = 11
	fsUI      float32 = 12
	fsRead    float32 = 14
	fsTitle   float32 = 17
	fsSection float32 = 20
	fsDisplay float32 = 24
)

const (
	readingWidth  float32 = 680
	settingsWidth float32 = 820
)

const (
	radiusPanel   float32 = 20
	radiusCard    float32 = 12
	radiusControl float32 = 8
	radiusBadge   float32 = 6
)

const (
	uiFont    = "SF Pro Text, Hiragino Sans GB, PingFang SC, system-ui, sans-serif"
	serifFont = "Source Han Serif SC VF, Source Han Serif SC, Songti SC, STSong, Georgia, serif"
	monoFont  = "SF Mono, Menlo, Consolas, monospace"
)

// composerGlass keeps Liquid Glass's rim and edge, but tints the large input
// with opaque paper so it does not bloom across the canvas.
func composerGlass(dark bool) glass.Glass {
	return glass.Glass{
		Style:   glass.Regular,
		Clarity: 0.5,
		Tint:    paletteFor(dark).composerTint,
	}
}

type palette struct {
	backdropA    ui.Color
	backdropB    ui.Color
	canvas       ui.Color
	composerTint ui.Color
	panelBg      ui.Color
	surface      ui.Color
	surface2     ui.Color
	text         ui.Color
	body         ui.Color
	muted        ui.Color
	faint        ui.Color
	hairline     ui.Color
	borderStrong ui.Color
	accent       ui.Color
	accentStrong ui.Color
	accentSoft   ui.Color
	recallBg     ui.Color
	accentLine   ui.Color
	rust         ui.Color
	green        ui.Color
	rustSoft     ui.Color
	greenSoft    ui.Color
	shadow       ui.Color
	navHover     ui.Color
	navLine      ui.Color
	fieldBg      ui.Color
}

func paletteFor(dark bool) palette {
	if dark {
		return palette{
			backdropA:    ui.RGBA(24, 25, 22, 0.97),
			backdropB:    ui.RGBA(33, 34, 29, 0.97),
			canvas:       ui.Hex("#171815"),
			composerTint: ui.Hex("#25261F"),
			panelBg:      ui.Hex("#25261F"),
			surface:      ui.RGBA(242, 240, 233, 0.035),
			surface2:     ui.RGBA(242, 240, 233, 0.06),
			text:         ui.Hex("#F2F0E9"),
			body:         ui.Hex("#D3D1C7"),
			muted:        ui.Hex("#A5A69B"),
			faint:        ui.Hex("#808177"),
			hairline:     ui.RGBA(242, 240, 233, 0.11),
			borderStrong: ui.RGBA(242, 240, 233, 0.20),
			accent:       ui.Hex("#93CDB8"),
			accentStrong: ui.Hex("#2F6B57"),
			accentSoft:   ui.RGBA(147, 205, 184, 0.14),
			recallBg:     ui.RGBA(147, 205, 184, 0.075),
			accentLine:   ui.RGBA(147, 205, 184, 0.38),
			rust:         ui.Hex("#FF8F78"),
			rustSoft:     ui.RGBA(255, 143, 120, 0.13),
			green:        ui.Hex("#79D49B"),
			greenSoft:    ui.RGBA(121, 212, 155, 0.11),
			shadow:       ui.RGBA(10, 12, 9, 0.55),
			navHover:     ui.RGBA(242, 240, 233, 0.09),
			navLine:      ui.RGBA(242, 240, 233, 0.08),
			fieldBg:      ui.RGBA(242, 240, 233, 0.07),
		}
	}
	return palette{
		backdropA:    ui.RGBA(247, 245, 239, 0.96),
		backdropB:    ui.RGBA(237, 235, 227, 0.96),
		canvas:       ui.Hex("#F3F1EB"),
		composerTint: ui.Hex("#FBFAF6"),
		panelBg:      ui.Hex("#FBFAF6"),
		surface:      ui.RGBA(255, 255, 255, 0.42),
		surface2:     ui.RGBA(255, 255, 255, 0.6),
		text:         ui.Hex("#1E1E1A"),
		body:         ui.Hex("#3C3C36"),
		muted:        ui.Hex("#6D6E64"),
		faint:        ui.Hex("#8E8F84"),
		hairline:     ui.RGBA(30, 30, 26, 0.10),
		borderStrong: ui.RGBA(30, 30, 26, 0.18),
		accent:       ui.Hex("#2F6B57"),
		accentStrong: ui.Hex("#275846"),
		accentSoft:   ui.RGBA(47, 107, 87, 0.12),
		recallBg:     ui.RGBA(47, 107, 87, 0.065),
		accentLine:   ui.RGBA(47, 107, 87, 0.36),
		rust:         ui.Hex("#B4472F"),
		rustSoft:     ui.RGBA(180, 71, 47, 0.09),
		green:        ui.Hex("#2B7A4B"),
		greenSoft:    ui.RGBA(43, 122, 75, 0.10),
		shadow:       ui.RGBA(39, 35, 25, 0.16),
		navHover:     ui.RGBA(30, 30, 26, 0.06),
		navLine:      ui.RGBA(30, 30, 26, 0.08),
		fieldBg:      ui.RGBA(255, 255, 255, 0.62),
	}
}

func appTheme(dark bool) *ui.Theme {
	t := ui.LightTheme()
	if dark {
		t = ui.DarkTheme()
	}
	p := paletteFor(dark)
	t.Background = p.canvas
	t.Surface = p.surface
	t.SurfaceHover = p.fieldBg
	t.SurfacePressed = p.hairline
	t.Border = p.hairline
	t.Text = p.text
	t.TextMuted = p.muted
	t.Accent = p.accent
	t.AccentText = ui.Hex("#FFFFFF")
	t.AccentHover = withAlpha(t.Accent, 0.9)
	t.AccentPressed = withAlpha(t.Accent, 0.8)
	t.Focus = p.accent
	t.Selection = p.accentSoft
	t.Danger = p.rust
	t.FontSize = fsUI
	t.Font = uiFont
	t.Radius = 10
	t.Spacing = 4
	t.Scrollbar = p.borderStrong
	t.ScrollbarWidth = 6
	return t
}

// heatStyle returns the fill, border and text colors of a mastery tile.
// Mastery reads as one jade scale so the heatmap follows the app's single
// accent instead of introducing another hue.
func heatStyle(mastery int, dark bool) (ui.Color, ui.Color, ui.Color) {
	if dark {
		switch mastery {
		case 1:
			return ui.RGBA(147, 205, 184, 0.26), ui.RGBA(147, 205, 184, 0.36), ui.Hex("#BDE5D7")
		case 2:
			return ui.RGBA(147, 205, 184, 0.36), ui.RGBA(147, 205, 184, 0.55), ui.Hex("#E6F7F1")
		case 3:
			return ui.Hex("#2F6B57"), ui.Hex("#93CDB8"), ui.Hex("#FFFFFF")
		default:
			return ui.RGBA(242, 240, 233, 0.15), ui.RGBA(242, 240, 233, 0.20), ui.Hex("#8A8B80")
		}
	}
	switch mastery {
	case 1:
		return ui.RGBA(47, 107, 87, 0.30), ui.RGBA(47, 107, 87, 0.42), ui.Hex("#2F6B57")
	case 2:
		return ui.RGBA(47, 107, 87, 0.44), ui.RGBA(47, 107, 87, 0.55), ui.Hex("#1F4E3E")
	case 3:
		return ui.Hex("#2F6B57"), ui.Hex("#275846"), ui.Hex("#FFFFFF")
	default:
		return ui.RGBA(30, 30, 26, 0.15), ui.RGBA(30, 30, 26, 0.20), ui.Hex("#8E8F84")
	}
}

func withAlpha(c ui.Color, alpha float32) ui.Color { return c.Alpha(alpha) }
