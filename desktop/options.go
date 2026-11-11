package main

import (
	"regexp"
	"strings"

	"github.com/egoist/mygo/ui"
)

// legacyOption is a choice line from an older plain-text quiz message. It is
// rendered like a quiz option, but has no answer state and is never clickable.
type legacyOption struct {
	Letter string
	Text   string
	Code   bool
}

var legacyOptionLine = regexp.MustCompile(
	`^\s*(?:[-*•]\s*)?(?:\*\*)?([A-Da-d])(?:\*\*)?\s*[.、．)]\s*(.+?)\s*$`,
)

// splitLegacyOptions finds a contiguous A/B/C[/D] choice list in an assistant
// message. The display layer uses this for messages written before save_quiz
// existed; the stored message is left untouched.
func splitLegacyOptions(text string) (before string, options []legacyOption, after string, ok bool) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if m := legacyOptionLine.FindStringSubmatch(line); m != nil &&
			strings.EqualFold(m[1], "A") {
			start = i
			break
		}
	}
	if start < 0 {
		return "", nil, "", false
	}

	want := byte('A')
	end := start
	for i := start; i < len(lines); i++ {
		line := lines[i]
		m := legacyOptionLine.FindStringSubmatch(line)
		if m == nil {
			// Some models wrap a long option onto indented continuation
			// lines; keep those with the option before stopping.
			if len(options) > 0 && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
				options[len(options)-1].Text += " " + strings.TrimSpace(line)
				end = i + 1
				continue
			}
			break
		}
		letter := strings.ToUpper(m[1])
		if letter != string(want) {
			break
		}
		raw := strings.TrimSpace(m[2])
		optionText := plainOptionText(raw)
		if optionText == "" {
			break
		}
		options = append(options, legacyOption{
			Letter: letter,
			Text:   optionText,
			Code:   strings.Contains(raw, "`") || looksLikeCode(optionText),
		})
		want++
		end = i + 1
		if want > 'D' {
			break
		}
	}
	if len(options) < 2 {
		return "", nil, "", false
	}
	before = strings.TrimSpace(strings.Join(lines[:start], "\n"))
	after = strings.TrimSpace(strings.Join(lines[end:], "\n"))
	return before, options, after, true
}

func plainOptionText(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	return strings.TrimSpace(s)
}

// optionContent draws one answer row. Quiz and legacy messages share it so
// the letter column, line height and padding stay identical.
func optionContent(c *ui.Context, p palette, letter, text string, code bool, fg, trailingColor ui.Color, trailing string) {
	ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Start).Children(func() {
		ui.Text(c, letter).Width(18).FontSize(fsCaption).Font(monoFont).
			TextColor(fg).Padding(1, 0, 0, 0)
		if code {
			ui.Text(c, text).Grow(1).MinWidth(0).FontSize(fsUI).Font(monoFont).
				LineHeight(1.6).TextColor(fg).MaxLines(6)
		} else {
			ui.Text(c, text).Grow(1).MinWidth(0).FontSize(fsRead).
				LineHeight(1.55).TextColor(fg).MaxLines(3)
		}
		if trailing != "" {
			ui.Text(c, trailing).FontSize(fsCaption).TextColor(trailingColor).
				Padding(1, 0, 0, 0)
		}
	})
}
