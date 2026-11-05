package main

import "github.com/egoist/mygo/ui"

// Type scale: 12 / 14 / 16 / 20 / 24 / 30.
// Compact editor scale: meta 11 / UI 12 / reading 13 / title 14 / section 16.
const (
	fsCaption float32 = 11
	fsUI      float32 = 12
	fsRead    float32 = 13
	fsTitle   float32 = 14
	fsSection float32 = 16
	fsDisplay float32 = 18
)

// Shadcn radii: rounded-lg 8 (cards), rounded-md 6 (controls), rounded-sm 4 (badges).
const (
	radiusCard    float32 = 8
	radiusControl float32 = 6
	radiusBadge   float32 = 4
)

const (
	uiFont    = "system-ui"
	serifFont = "Songti SC, Noto Serif SC, Source Han Serif SC, Georgia, serif"
	monoFont  = "SF Mono, Menlo, Consolas, monospace"
)

type palette struct {
	canvas         ui.Color
	surface        ui.Color
	surface2       ui.Color
	sidebar        ui.Color
	sidebarText    ui.Color
	sidebarMuted   ui.Color
	text           ui.Color
	body           ui.Color
	muted          ui.Color
	faint          ui.Color
	hairline       ui.Color
	borderStrong   ui.Color
	blue           ui.Color
	blueSoft       ui.Color
	recallBg       ui.Color
	blueLine       ui.Color
	rust           ui.Color
	green          ui.Color
	rustSoft       ui.Color
	shadow         ui.Color
	navActive      ui.Color
	navHover       ui.Color
	navLine        ui.Color
	fieldBg        ui.Color
	composerBg     ui.Color
	composerText   ui.Color
	composerMuted  ui.Color
	composerLine   ui.Color
	composerHover  ui.Color
	composerCircle ui.Color
	composerInk    ui.Color
	composerDanger ui.Color
}

func paletteFor(dark bool) palette {
	if dark {
		return palette{
			canvas:         ui.Hex("#1A1A1A"),
			surface:        ui.Hex("#1F1F1F"),
			surface2:       ui.Hex("#242424"),
			sidebar:        ui.Hex("#1A1A1A"),
			sidebarText:    ui.Hex("#F2F2F2"),
			sidebarMuted:   ui.Hex("#9A9A9A"),
			text:           ui.Hex("#F2F2F2"),
			body:           ui.Hex("#D4D4D4"),
			muted:          ui.Hex("#A3A3A3"),
			faint:          ui.Hex("#8A8A8A"),
			hairline:       ui.Hex("#303030"),
			borderStrong:   ui.Hex("#4A4A4A"),
			blue:           ui.Hex("#F59E0B"),
			blueSoft:       ui.Hex("#3A2A14"),
			recallBg:       ui.Hex("#2A2118"),
			blueLine:       ui.Hex("#5A3E1B"),
			rust:           ui.Hex("#F97066"),
			green:          ui.Hex("#4ADE80"),
			rustSoft:       ui.Hex("#3A1D1A"),
			shadow:         ui.RGBA(0, 0, 0, 0.45),
			navActive:      ui.RGBA(255, 255, 255, 0.07),
			navHover:       ui.RGBA(255, 255, 255, 0.04),
			navLine:        ui.RGBA(255, 255, 255, 0.08),
			fieldBg:        ui.Hex("#262626"),
			composerBg:     ui.Hex("#242424"),
			composerText:   ui.Hex("#F2F2F2"),
			composerMuted:  ui.Hex("#9A9A9A"),
			composerLine:   ui.RGBA(255, 255, 255, 0.12),
			composerHover:  ui.RGBA(255, 255, 255, 0.08),
			composerCircle: ui.Hex("#F5F5F5"),
			composerInk:    ui.Hex("#171717"),
			composerDanger: ui.Hex("#F97066"),
		}
	}
	return palette{
		canvas:         ui.Hex("#F5F5F5"),
		surface:        ui.Hex("#FFFFFF"),
		surface2:       ui.Hex("#FAFAFA"),
		sidebar:        ui.Hex("#F5F5F5"),
		sidebarText:    ui.Hex("#1A1A1A"),
		sidebarMuted:   ui.Hex("#7A7A7A"),
		text:           ui.Hex("#1A1A1A"),
		body:           ui.Hex("#3A3A3A"),
		muted:          ui.Hex("#6F6F6F"),
		faint:          ui.Hex("#8C8C8C"),
		hairline:       ui.Hex("#E6E6E6"),
		borderStrong:   ui.Hex("#D2D2D2"),
		blue:           ui.Hex("#B45309"),
		blueSoft:       ui.Hex("#FDF4EA"),
		blueLine:       ui.Hex("#EAD5BF"),
		rust:           ui.Hex("#B42318"),
		green:          ui.Hex("#16A34A"),
		rustSoft:       ui.Hex("#FEF3F2"),
		shadow:         ui.RGBA(17, 17, 17, 0.05),
		navActive:      ui.Hex("#ECECEC"),
		navHover:       ui.Hex("#F1F1F1"),
		navLine:        ui.Hex("#E6E6E6"),
		fieldBg:        ui.Hex("#F5F5F5"),
		composerBg:     ui.Hex("#FFFFFF"),
		composerText:   ui.Hex("#1A1A1A"),
		composerMuted:  ui.Hex("#6F6F6F"),
		composerLine:   ui.Hex("#E6E6E6"),
		composerHover:  ui.Hex("#F5F5F5"),
		composerCircle: ui.Hex("#1A1A1A"),
		composerInk:    ui.Hex("#FFFFFF"),
		composerDanger: ui.Hex("#B42318"),
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
	t.Accent = p.text
	t.AccentText = p.canvas
	if dark {
		t.Accent = p.text
		t.AccentText = ui.Hex("#111110")
	} else {
		t.Accent = p.text
		t.AccentText = ui.Hex("#F7F6F1")
	}
	t.AccentHover = p.text
	t.AccentPressed = p.text
	t.Focus = p.blue
	t.Selection = p.blueSoft
	t.Danger = p.rust
	t.FontSize = fsUI
	t.Font = uiFont
	t.Radius = 8
	t.Spacing = 4
	t.Scrollbar = p.borderStrong
	t.ScrollbarWidth = 6
	return t
}

func serif(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).Font(serifFont)
}

// heatStyle returns the fill, border and text colors of a mastery tile.
func heatStyle(mastery int, dark bool) (ui.Color, ui.Color, ui.Color) {
	p := paletteFor(dark)
	switch mastery {
	case 1:
		if dark {
			return ui.Hex("#2E2317"), ui.Hex("#4A3824"), ui.Hex("#E8C9A8")
		}
		return ui.Hex("#FDF4EA"), ui.Hex("#EAD5BF"), ui.Hex("#8A4B16")
	case 2:
		if dark {
			return ui.Hex("#5A3B1F"), ui.Hex("#7A522A"), ui.Hex("#F5E0C8")
		}
		return ui.Hex("#F5D9BF"), ui.Hex("#E7C09A"), ui.Hex("#6B3A10")
	case 3:
		if dark {
			return ui.Hex("#D97706"), ui.Hex("#D97706"), ui.Hex("#FFFFFF")
		}
		return p.blue, p.blue, ui.Hex("#FFFFFF")
	default:
		if dark {
			return ui.Hex("#242424"), ui.Hex("#333333"), ui.Hex("#8A8A8A")
		}
		return ui.Hex("#F5F5F5"), ui.Hex("#E6E6E6"), ui.Hex("#8C8C8C")
	}
}

func withAlpha(c ui.Color, alpha float32) ui.Color { return c.Alpha(alpha) }

// composerTheme keeps the caret and selection legible on the composer surface.
func composerTheme(base *ui.Theme) *ui.Theme {
	p := paletteFor(base.Dark)
	t := appTheme(base.Dark)
	t.Background = ui.Transparent
	t.Surface = ui.Transparent
	t.SurfaceHover = ui.Transparent
	t.SurfacePressed = ui.Transparent
	t.Border = p.composerLine
	t.Text = p.composerText
	t.TextMuted = p.composerMuted
	t.Accent = p.composerText
	t.AccentText = p.composerInk
	t.Focus = p.composerText
	t.Selection = withAlpha(p.composerText, 0.24)
	t.Danger = p.composerDanger
	t.FontSize = 14
	return t
}
