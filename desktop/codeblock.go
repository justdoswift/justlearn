package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

const codePreviewLines = 9

// codeBlock is the shared code surface for lessons and generated canvases.
// Long snippets start folded; the reader can expand them in place.
func (a *app) codeBlock(c *ui.Context, p palette, key, language, code, caption string) {
	code = strings.Trim(strings.TrimSpace(code), "\n")
	if code == "" {
		return
	}
	lines := strings.Split(code, "\n")
	expanded := a.expandedCode[key]
	shown := code
	if len(lines) > codePreviewLines && !expanded {
		shown = strings.Join(lines[:codePreviewLines], "\n")
	}
	if language == "" {
		language = languageForCode(code)
	}
	ui.Column(c).FillWidth().Background(p.surface2).BorderWidth(1).BorderColor(p.hairline).
		Radius(radiusControl).Clip().Padding(9, 11).Gap(6).Children(func() {
		if language != "" || caption != "" {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				if language != "" {
					ui.Text(c, strings.ToUpper(language)).FontSize(fsCaption).Font(monoFont).
						FontWeight(400).TextColor(p.muted)
				}
				ui.Spacer(c)
				if caption != "" {
					ui.Text(c, caption).FontSize(fsCaption).TextColor(p.muted).MaxLines(1)
				}
			})
		}
		ui.ScrollHorizontal(c).FillWidth().Children(func() {
			ui.Text(c, shown).FontSize(fsUI).Font(monoFont).LineHeight(1.65).TextColor(p.body)
		})
		if len(lines) > codePreviewLines {
			label := "展开全部 " + strconv.Itoa(len(lines)) + " 行"
			if expanded {
				label = "收起代码"
			}
			more := ui.ButtonBase(c).Height(26).PaddingX(10).Radius(radiusBadge).
				AlignItems(ui.Center).Transition(ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond}).
				Children(func() {
					ui.Text(c, label).FontSize(fsCaption).FontWeight(400).TextColor(p.muted)
				})
			if more.Hovered() {
				more.Background(p.navHover)
			}
			if more.Clicked() {
				a.expandedCode[key] = !expanded
			}
		}
	})
}

func looksLikeCode(s string) bool {
	low := strings.ToLower(s)
	if strings.Contains(low, "\n") {
		return true
	}
	for _, marker := range []string{"select ", "select\n", "group by", "count(", "join ", "def ", "func ", "package ", "import "} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return strings.Contains(low, " from ") && strings.Contains(low, " where ")
}

func languageForCode(code string) string {
	low := strings.ToLower(code)
	switch {
	case strings.Contains(low, "select ") || strings.Contains(low, "from ") ||
		strings.Contains(low, "group by") || strings.Contains(low, "where "):
		return "sql"
	case strings.Contains(low, "package ") || strings.Contains(low, "func "):
		return "go"
	case strings.Contains(low, "def ") || strings.Contains(low, "import "):
		return "python"
	default:
		return ""
	}
}
