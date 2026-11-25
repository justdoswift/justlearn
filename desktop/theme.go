package main

import (
	"time"

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

// Tracking is size-specific, never one value for every size: display text
// reads too loose as it grows and wants negative tracking, small text wants
// a little positive tracking to stay legible. CJK sits near zero at reading
// sizes, so only the Latin display and the tiny mono labels move.
const (
	trackDisplay float32 = -0.3
	trackTitle   float32 = -0.1
	trackBody    float32 = 0
	trackUI      float32 = 0.1
	trackCaption float32 = 0.2
	trackMono    float32 = 0.5
)

// typeStyle is one step of the scale: size, weight, leading and tracking
// chosen together. Hierarchy comes from the set, not from size alone.
type typeStyle struct {
	Size     float32
	Weight   int
	Leading  float32
	Tracking float32
	Font     string
}

// apply writes the style onto an element. Zero fields are left alone, so a
// style can inherit the theme's family or weight.
func (s typeStyle) apply(el ui.Element) ui.Element {
	el.FontSize(s.Size)
	if s.Weight > 0 {
		el.FontWeight(s.Weight)
	}
	if s.Leading > 0 {
		el.LineHeight(s.Leading)
	}
	if s.Tracking != 0 {
		el.LetterSpacing(s.Tracking)
	}
	if s.Font != "" {
		el.Font(s.Font)
	}
	return el
}

var (
	// The lesson title: large, tight leading, negative tracking.
	styleDisplay = typeStyle{Size: 21, Weight: 400, Leading: 1.3, Tracking: trackDisplay, Font: serifFont}
	// Section and panel headings.
	styleTitle = typeStyle{Size: fsTitle, Weight: 400, Leading: 1.38, Tracking: trackTitle}
	// Reading copy.
	styleBody = typeStyle{Size: fsRead, Weight: 400, Leading: 1.72, Tracking: trackBody}
	// Controls and labels.
	styleUI = typeStyle{Size: fsUI, Weight: 400, Leading: 1.4, Tracking: trackUI}
	// Secondary meta.
	styleCaption = typeStyle{Size: fsCaption, Weight: 400, Leading: 1.4, Tracking: trackCaption}
	// Small mono labels: stage numbers, counters, kickers.
	styleMono = typeStyle{Size: 9, Weight: 400, Leading: 1.2, Tracking: trackMono, Font: monoFont}
)

// Motion. Apple's defaults are critically damped: nothing overshoots unless
// a gesture threw it. Bounce is reserved for momentum, so every spring here
// settles without it.
var (
	// springSheet brings a panel or sheet in. Response 0.34, damping 1.0.
	springSheet = ui.Animation{Spring: ui.SpringOptions{Duration: 340 * time.Millisecond}}
	// springMove repositions something that is already on screen.
	springMove = ui.Animation{Spring: ui.SpringOptions{Duration: 400 * time.Millisecond}}
	// springHover answers the pointer: quick, no overshoot.
	springHover = ui.Animation{Spring: ui.SpringOptions{Duration: 260 * time.Millisecond}}
	// springPress is the press and release of a control.
	springPress = ui.Animation{Spring: ui.SpringOptions{Duration: 220 * time.Millisecond}}
)

// pressPose is the press feedback: the control acknowledges the pointer the
// instant it goes down, not when it comes up. Hover is a whisper on top.
func pressPose(el ui.Element, hovered, pressed bool, base float32) {
	pose := ui.Pose{Scale: base}
	switch {
	case pressed:
		pose.Scale = base * 0.97
	case hovered:
		pose.Scale = base * 1.02
	}
	el.AnimatePose(pose, springPress)
}

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

// cardGlass is the floating surface shared by the route preview and the
// ability detail: the same material, so the two read as one layer of chrome.
func cardGlass(dark bool, alpha float32) glass.Glass {
	p := paletteFor(dark)
	return glass.Glass{
		Style:   glass.Regular,
		Clarity: 0.34,
		Tint:    withAlpha(p.panelBg, alpha),
	}
}

// composerShadow is the lift under the composer. A raised surface casts a
// soft, wide shadow; in light mode the paper needs a little more of it to
// separate from a canvas of almost the same value.
func composerShadow(dark bool) ui.Color {
	p := paletteFor(dark)
	if dark {
		return ui.RGBA(0, 0, 0, 0.10)
	}
	return withAlpha(p.shadow, 0.55)
}

// composerEdge is the brightness of the lit top edge. In light mode the
// material is already near-white, so the highlight has to be a warm tint
// rather than pure white or it disappears into the fill.
func composerEdge(dark bool) float32 {
	if dark {
		return 0.22
	}
	return 0.9
}

// cardShadow is the lift under a floating card. Bigger surfaces read as
// thicker, so the shadow is deeper than the composer's.
func cardShadow(dark bool) ui.Color {
	p := paletteFor(dark)
	if dark {
		return withAlpha(p.shadow, 0.85)
	}
	return withAlpha(p.shadow, 0.75)
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
			backdropA:    ui.RGBA(22, 23, 20, 0.97),
			backdropB:    ui.RGBA(29, 30, 26, 0.97),
			canvas:       ui.Hex("#171815"),
			composerTint: ui.Hex("#2C2D26"),
			panelBg:      ui.Hex("#2A2B24"),
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
		backdropA:    ui.Hex("#F4F1E9"),
		backdropB:    ui.Hex("#EDE9DC"),
		canvas:       ui.Hex("#F3F1EB"),
		// The composer and panels sit above the canvas, so they must be
		// lighter than it. A tint a shade off the canvas reads as paint,
		// not as a raised material.
		composerTint: ui.Hex("#FCFAF3"),
		panelBg:      ui.Hex("#FCFAF4"),
		surface:      ui.RGBA(255, 251, 241, 0.34),
		surface2:     ui.RGBA(255, 251, 241, 0.54),
		text:         ui.Hex("#1E1E1A"),
		body:         ui.Hex("#3C3C36"),
		muted:        ui.Hex("#6D6E64"),
		faint:        ui.Hex("#8E8F84"),
		hairline:     ui.RGBA(60, 52, 35, 0.12),
		borderStrong: ui.RGBA(60, 52, 35, 0.20),
		accent:       ui.Hex("#2F6B57"),
		accentStrong: ui.Hex("#275846"),
		accentSoft:   ui.RGBA(47, 107, 87, 0.10),
		recallBg:     ui.RGBA(47, 107, 87, 0.06),
		accentLine:   ui.RGBA(47, 107, 87, 0.34),
		rust:         ui.Hex("#B4472F"),
		rustSoft:     ui.RGBA(180, 71, 47, 0.09),
		green:        ui.Hex("#2B7A4B"),
		greenSoft:    ui.RGBA(43, 122, 75, 0.10),
		shadow:       ui.RGBA(64, 55, 38, 0.16),
		navHover:     ui.RGBA(60, 52, 35, 0.06),
		navLine:      ui.RGBA(60, 52, 35, 0.08),
		fieldBg:      ui.RGBA(255, 251, 241, 0.72),
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
// accent instead of introducing another hue. The lowest step stays visible
// on warm paper: a mark at 10% alpha disappeared into the canvas and the
// profile read as empty rather than as "not yet learned".
func heatStyle(mastery int, dark bool) (ui.Color, ui.Color, ui.Color) {
	if dark {
		switch mastery {
		case 1:
			return ui.RGBA(147, 205, 184, 0.42), ui.RGBA(147, 205, 184, 0.36), ui.Hex("#BDE5D7")
		case 2:
			return ui.RGBA(147, 205, 184, 0.66), ui.RGBA(147, 205, 184, 0.55), ui.Hex("#E6F7F1")
		case 3:
			return ui.Hex("#93CDB8"), ui.Hex("#93CDB8"), ui.Hex("#171815")
		default:
			return ui.RGBA(242, 240, 233, 0.26), ui.RGBA(242, 240, 233, 0.20), ui.Hex("#8A8B80")
		}
	}
	switch mastery {
	case 1:
		return ui.RGBA(47, 107, 87, 0.34), ui.RGBA(47, 107, 87, 0.42), ui.Hex("#2F6B57")
	case 2:
		return ui.RGBA(47, 107, 87, 0.58), ui.RGBA(47, 107, 87, 0.55), ui.Hex("#1F4E3E")
	case 3:
		return ui.Hex("#2F6B57"), ui.Hex("#275846"), ui.Hex("#FFFFFF")
	default:
		return ui.RGBA(60, 52, 35, 0.22), ui.RGBA(47, 107, 87, 0.24), ui.Hex("#5F7A6E")
	}
}

func withAlpha(c ui.Color, alpha float32) ui.Color { return c.Alpha(alpha) }
