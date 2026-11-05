package main

import "github.com/egoist/mygo/ui"

// Type scale: meta 11 / UI 12 / reading 14 / title 17 / section 20 / display 24.
const (
	fsCaption float32 = 11
	fsUI      float32 = 12
	fsRead    float32 = 14
	fsTitle   float32 = 17
	fsSection float32 = 20
	fsDisplay float32 = 24
)

const serifFont = "Songti SC, Noto Serif SC, Source Han Serif SC, Georgia, serif"

func serif(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).Font(serifFont)
}

// Shadcn radii: rounded-lg 8 (cards), rounded-md 6 (controls), rounded-sm 4 (badges).
const (
	radiusCard    float32 = 8
	radiusControl float32 = 6
	radiusBadge   float32 = 4
)

const (
	uiFont   = "system-ui"
	monoFont = "SF Mono, Menlo, Consolas, monospace"
)

type palette struct {
	canvas       ui.Color
	surface      ui.Color
	surface2     ui.Color
	sidebar      ui.Color
	sidebarText  ui.Color
	sidebarMuted ui.Color
	text         ui.Color
	body         ui.Color
	muted        ui.Color
	faint        ui.Color
	hairline     ui.Color
	borderStrong ui.Color
	blue         ui.Color
	blueStrong   ui.Color
	blueSoft     ui.Color
	recallBg     ui.Color
	blueLine     ui.Color
	rust         ui.Color
	green        ui.Color
	rustSoft     ui.Color
	shadow       ui.Color
	navHover     ui.Color
	navLine      ui.Color
	fieldBg      ui.Color
}

func paletteFor(dark bool) palette {
	if dark {
		return palette{
			canvas:       ui.Hex("#1E1E1E"),
			surface:      ui.Hex("#1E1E1E"),
			surface2:     ui.Hex("#2D2D2E"),
			sidebar:      ui.Hex("#252526"),
			sidebarText:  ui.Hex("#E8E8EA"),
			sidebarMuted: ui.Hex("#9A9A9E"),
			text:         ui.Hex("#E8E8EA"),
			body:         ui.Hex("#CFCFD2"),
			muted:        ui.Hex("#A0A0A5"),
			faint:        ui.Hex("#7C7C82"),
			hairline:     ui.Hex("#3A3A3C"),
			borderStrong: ui.Hex("#4C4C50"),
			blue:         ui.Hex("#0A84FF"),
			blueStrong:   ui.Hex("#0A6CD6"),
			blueSoft:     ui.Hex("#0B2E4F"),
			recallBg:     ui.Hex("#102A43"),
			blueLine:     ui.Hex("#1F4E7A"),
			rust:         ui.Hex("#FF453A"),
			rustSoft:     ui.Hex("#3A1C1B"),
			green:        ui.Hex("#30D158"),
			shadow:       ui.RGBA(0, 0, 0, 0.55),
			navHover:     ui.Hex("#2F2F31"),
			navLine:      ui.Hex("#3A3A3C"),
			fieldBg:      ui.Hex("#2C2C2E"),
		}
	}
	return palette{
		canvas:       ui.Hex("#F5F5F5"),
		surface:      ui.Hex("#FFFFFF"),
		surface2:     ui.Hex("#FAFAFA"),
		sidebar:      ui.Hex("#F5F5F5"),
		sidebarText:  ui.Hex("#1A1A1A"),
		sidebarMuted: ui.Hex("#7A7A7A"),
		text:         ui.Hex("#1A1A1A"),
		body:         ui.Hex("#3A3A3A"),
		muted:        ui.Hex("#6F6F6F"),
		faint:        ui.Hex("#8C8C8C"),
		hairline:     ui.Hex("#E6E6E6"),
		borderStrong: ui.Hex("#D2D2D2"),
		blue:         ui.Hex("#C2410C"),
		blueStrong:   ui.Hex("#9A3412"),
		blueSoft:     ui.Hex("#FFF4ED"),
		blueLine:     ui.Hex("#FED7AA"),
		rust:         ui.Hex("#B42318"),
		green:        ui.Hex("#16A34A"),
		rustSoft:     ui.Hex("#FEF3F2"),
		shadow:       ui.RGBA(17, 17, 17, 0.05),
		navHover:     ui.Hex("#F1F1F1"),
		navLine:      ui.Hex("#E6E6E6"),
		fieldBg:      ui.Hex("#F5F5F5"),
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
		t.Accent = p.blue
		t.AccentText = ui.Hex("#FFFFFF")
	} else {
		t.Accent = p.text
		t.AccentText = ui.Hex("#F7F6F1")
	}
	t.AccentHover = withAlpha(t.Accent, 0.9)
	t.AccentPressed = withAlpha(t.Accent, 0.8)
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

// heatStyle returns the fill, border and text colors of a mastery tile.
func heatStyle(mastery int, dark bool) (ui.Color, ui.Color, ui.Color) {
	p := paletteFor(dark)
	switch mastery {
	case 1:
		if dark {
			return ui.Hex("#442A1B"), ui.Hex("#6B4226"), ui.Hex("#F8C9A8")
		}
		return ui.Hex("#FCE7D6"), ui.Hex("#F5C7A5"), ui.Hex("#9A3412")
	case 2:
		if dark {
			return ui.Hex("#8A4B24"), ui.Hex("#A65D2E"), ui.Hex("#FBE0C8")
		}
		return ui.Hex("#F5B48A"), ui.Hex("#EA9A6A"), ui.Hex("#7C2D12")
	case 3:
		if dark {
			return ui.Hex("#EA580C"), ui.Hex("#EA580C"), ui.Hex("#FFFFFF")
		}
		return p.blue, p.blue, ui.Hex("#FFFFFF")
	default:
		if dark {
			return ui.Hex("#242424"), ui.Hex("#333333"), ui.Hex("#8A8A8A")
		}
		return ui.Hex("#ECECEC"), ui.Hex("#DCDCDC"), ui.Hex("#8C8C8C")
	}
}

func withAlpha(c ui.Color, alpha float32) ui.Color { return c.Alpha(alpha) }
