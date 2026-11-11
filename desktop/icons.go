package main

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// lineIcon parses a 24×24 stroked icon drawn in currentColor.
func lineIcon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	iconSidebar = lineIcon(`<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/>`)
	iconConfig  = lineIcon(`<path d="M4 7h8M18 7h2M4 17h2M12 17h8"/><circle cx="15" cy="7" r="2.2"/><circle cx="8" cy="17" r="2.2"/>`)
	iconPlus    = lineIcon(`<path d="M12 5v14M5 12h14"/>`)
	iconArrowUp = lineIcon(`<path d="M12 19V5M6 11l6-6 6 6"/>`)
	iconStop    = lineIcon(`<rect x="7" y="7" width="10" height="10" rx="2" fill="currentColor" stroke="none"/>`)
	iconClose   = lineIcon(`<path d="M6 6l12 12M18 6 6 18"/>`)
	iconChevron = lineIcon(`<path d="m6 9 6 6 6-6"/>`)
	iconCheck   = lineIcon(`<path d="m5 12 4 4L19 6"/>`)
	iconAuto    = lineIcon(`<circle cx="12" cy="12" r="8"/><path d="M12 4v16"/>`)
	iconSun     = lineIcon(`<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>`)
	iconMoon    = lineIcon(`<path d="M20 14.5A8.5 8.5 0 0 1 9.5 4a8.5 8.5 0 1 0 10.5 10.5z"/>`)
	iconInfo    = lineIcon(`<circle cx="12" cy="12" r="9"/><path d="M12 11v5M12 8h.01"/>`)
)

// iconButton builds a quiet 28×28 toolbar control.
func iconButton(c *ui.Context, p palette, icon *ui.SVG, label string) ui.Element {
	btn := ui.ButtonBase(c).Size(28, 28).Radius(radiusControl).Label(label).Tooltip(label).Center().
		Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond})
	btn.Children(func() {
		ui.Icon(c, icon).Size(15, 15).TextColor(p.muted)
	})
	if btn.Hovered() {
		btn.Background(p.navHover)
	}
	return btn
}
