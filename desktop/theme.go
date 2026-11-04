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
			canvas:         ui.Hex("#121211"),
			surface:        ui.Hex("#1B1B19"),
			surface2:       ui.Hex("#171715"),
			sidebar:        ui.Hex("#0C0C0B"),
			sidebarText:    ui.Hex("#F2F1EB"),
			sidebarMuted:   ui.Hex("#9A978E"),
			text:           ui.Hex("#F2F1EB"),
			body:           ui.Hex("#DAD8D0"),
			muted:          ui.Hex("#A09D94"),
			faint:          ui.Hex("#8C8981"),
			hairline:       ui.Hex("#2C2C28"),
			borderStrong:   ui.Hex("#3B3A34"),
			blue:           ui.Hex("#8FA3FF"),
			blueSoft:       ui.Hex("#1B2347"),
			recallBg:       ui.Hex("#1A1B1F"),
			blueLine:       ui.Hex("#33459E"),
			rust:           ui.Hex("#E08A7A"),
			rustSoft:       ui.Hex("#3A211D"),
			shadow:         ui.RGBA(0, 0, 0, 0.45),
			navActive:      ui.RGBA(255, 255, 255, 0.07),
			navHover:       ui.RGBA(255, 255, 255, 0.04),
			navLine:        ui.RGBA(255, 255, 255, 0.08),
			fieldBg:        ui.Hex("#21211E"),
			composerBg:     ui.Hex("#2C2C28"),
			composerText:   ui.Hex("#F2F0EA"),
			composerMuted:  ui.Hex("#9C9991"),
			composerLine:   ui.RGBA(255, 255, 255, 0.12),
			composerHover:  ui.RGBA(255, 255, 255, 0.10),
			composerCircle: ui.Hex("#F7F6F1"),
			composerInk:    ui.Hex("#1B1B18"),
			composerDanger: ui.Hex("#E08A7A"),
		}
	}
	return palette{
		canvas:         ui.Hex("#FAFAF7"),
		surface:        ui.Hex("#FFFFFF"),
		surface2:       ui.Hex("#F5F5F1"),
		sidebar:        ui.Hex("#F7F7F3"),
		sidebarText:    ui.Hex("#171714"),
		sidebarMuted:   ui.Hex("#73736C"),
		text:           ui.Hex("#141413"),
		body:           ui.Hex("#2C2B27"),
		muted:          ui.Hex("#6C6A62"),
		faint:          ui.Hex("#726F67"),
		hairline:       ui.Hex("#E8E7E1"),
		borderStrong:   ui.Hex("#D5D4CC"),
		blue:           ui.Hex("#2B4EDB"),
		blueSoft:       ui.Hex("#EDF0FF"),
		blueLine:       ui.Hex("#C9D3FF"),
		rust:           ui.Hex("#A33D2B"),
		rustSoft:       ui.Hex("#FBEDE9"),
		shadow:         ui.RGBA(17, 17, 17, 0.06),
		navActive:      ui.Hex("#ECECE6"),
		navHover:       ui.Hex("#F0F0EA"),
		navLine:        ui.Hex("#E7E6E0"),
		fieldBg:        ui.Hex("#F5F5F1"),
		composerBg:     ui.Hex("#FFFFFF"),
		composerText:   ui.Hex("#141413"),
		composerMuted:  ui.Hex("#6C6A62"),
		composerLine:   ui.Hex("#DCDAD2"),
		composerHover:  ui.Hex("#F0F0EA"),
		composerCircle: ui.Hex("#1B1B18"),
		composerInk:    ui.Hex("#F7F6F1"),
		composerDanger: ui.Hex("#A33D2B"),
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
			return ui.Hex("#1C2340"), ui.Hex("#2E3A66"), ui.Hex("#C6D0FF")
		}
		return ui.Hex("#D6DFFF"), ui.Hex("#BFCBFF"), ui.Hex("#24306E")
	case 2:
		if dark {
			return ui.Hex("#31439B"), ui.Hex("#4658B8"), ui.Hex("#E6EBFF")
		}
		return ui.Hex("#9FB0FA"), ui.Hex("#8497F0"), ui.Hex("#15246B")
	case 3:
		if dark {
			return ui.Hex("#3B51C9"), ui.Hex("#5A6FE0"), ui.Hex("#FFFFFF")
		}
		return p.blue, p.blue, ui.Hex("#FFFFFF")
	default:
		if dark {
			return ui.Hex("#201F1C"), ui.Hex("#33322D"), ui.Hex("#8C8981")
		}
		return ui.Hex("#ECEAE3"), ui.Hex("#DCD8CD"), ui.Hex("#726F67")
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
