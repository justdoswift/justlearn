// Package genui defines the small, typed vocabulary the tutor can render on
// the chat canvas. The model never sends HTML: it sends these blocks, and the
// desktop renders them with the app's own components.
package genui

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Payload is one canvas message.
type Payload struct {
	NodeID string  `json:"nodeId,omitempty"`
	Title  string  `json:"title,omitempty"`
	Blocks []Block `json:"blocks"`
}

// Block is one unit on the canvas. Type decides which fields are used.
type Block struct {
	Type string `json:"type"`

	Text    string `json:"text,omitempty"`
	Tone    string `json:"tone,omitempty"`
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
	Caption string `json:"caption,omitempty"`

	Columns []string   `json:"columns,omitempty"`
	Rows    [][]string `json:"rows,omitempty"`
	Items   []Item     `json:"items,omitempty"`

	Left  *Side `json:"left,omitempty"`
	Right *Side `json:"right,omitempty"`

	Language string `json:"language,omitempty"`
	Code     string `json:"code,omitempty"`

	Cards  []Card    `json:"cards,omitempty"`
	Labels []string  `json:"labels,omitempty"`
	Values []float64 `json:"values,omitempty"`
	Unit   string    `json:"unit,omitempty"`
}

type Item struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type Side struct {
	Title string   `json:"title,omitempty"`
	Items []string `json:"items,omitempty"`
}

type Card struct {
	Front string `json:"front,omitempty"`
	Back  string `json:"back,omitempty"`
}

const (
	MaxBlocks = 6
	MaxRows   = 12
	MaxCols   = 6
)

// Normalize trims the payload, drops blocks the renderer cannot use, and
// bounds the sizes so a model mistake cannot make the canvas unbounded.
func Normalize(p Payload) (Payload, error) {
	p.Title = trim(p.Title, 80)
	blocks := make([]Block, 0, len(p.Blocks))
	for _, b := range p.Blocks {
		if len(blocks) >= MaxBlocks {
			break
		}
		if nb, ok := normalizeBlock(b); ok {
			blocks = append(blocks, nb)
		}
	}
	if len(blocks) == 0 {
		return Payload{}, errors.New("画布没有可渲染的内容；至少给一个有效 block")
	}
	p.Blocks = blocks
	return p, nil
}

func normalizeBlock(b Block) (Block, bool) {
	b.Type = strings.ToLower(strings.TrimSpace(b.Type))
	b.Text = strings.TrimSpace(b.Text)
	b.Title = trim(b.Title, 80)
	b.Body = strings.TrimSpace(b.Body)
	b.Caption = trim(b.Caption, 120)
	switch b.Type {
	case "text":
		if b.Text == "" {
			return Block{}, false
		}
		return b, true
	case "callout":
		if b.Body == "" {
			return Block{}, false
		}
		b.Tone = strings.ToLower(strings.TrimSpace(b.Tone))
		switch b.Tone {
		case "info", "success", "warning":
		default:
			b.Tone = "info"
		}
		return b, true
	case "quote":
		if b.Text == "" {
			return Block{}, false
		}
		return b, true
	case "steps":
		items := make([]Item, 0, len(b.Items))
		for _, item := range b.Items {
			item.Title = strings.TrimSpace(item.Title)
			item.Body = strings.TrimSpace(item.Body)
			if item.Title == "" && item.Body == "" {
				continue
			}
			items = append(items, item)
			if len(items) == 6 {
				break
			}
		}
		if len(items) < 2 {
			return Block{}, false
		}
		b.Items = items
		return b, true
	case "table":
		if len(b.Columns) < 2 {
			return Block{}, false
		}
		cols := make([]string, 0, min(len(b.Columns), MaxCols))
		for _, col := range b.Columns {
			cols = append(cols, strings.TrimSpace(col))
			if len(cols) == MaxCols {
				break
			}
		}
		rows := make([][]string, 0, min(len(b.Rows), MaxRows))
		for _, row := range b.Rows {
			cells := make([]string, len(cols))
			for i := range cols {
				if i < len(row) {
					cells[i] = strings.TrimSpace(row[i])
				}
			}
			rows = append(rows, cells)
			if len(rows) == MaxRows {
				break
			}
		}
		if len(rows) == 0 {
			return Block{}, false
		}
		b.Columns, b.Rows = cols, rows
		return b, true
	case "compare":
		left, right := b.Left, b.Right
		if left == nil || right == nil {
			return Block{}, false
		}
		if nl, ok := normalizeSide(*left); ok {
			b.Left = &nl
		} else {
			return Block{}, false
		}
		if nr, ok := normalizeSide(*right); ok {
			b.Right = &nr
		} else {
			return Block{}, false
		}
		return b, true
	case "code":
		b.Code = strings.TrimSpace(b.Code)
		if b.Code == "" {
			return Block{}, false
		}
		b.Language = trim(b.Language, 24)
		if runes := []rune(b.Code); len(runes) > 4000 {
			b.Code = string(runes[:4000])
		}
		return b, true
	case "flashcards":
		cards := make([]Card, 0, len(b.Cards))
		for _, card := range b.Cards {
			card.Front = strings.TrimSpace(card.Front)
			card.Back = strings.TrimSpace(card.Back)
			if card.Front == "" || card.Back == "" {
				continue
			}
			cards = append(cards, card)
			if len(cards) == 8 {
				break
			}
		}
		if len(cards) < 2 {
			return Block{}, false
		}
		b.Cards = cards
		return b, true
	case "bar":
		n := min(len(b.Labels), len(b.Values))
		if n < 2 {
			return Block{}, false
		}
		if n > 8 {
			n = 8
		}
		labels := make([]string, n)
		values := make([]float64, n)
		for i := 0; i < n; i++ {
			labels[i] = strings.TrimSpace(b.Labels[i])
			if values[i] = b.Values[i]; math.IsNaN(values[i]) || math.IsInf(values[i], 0) || values[i] < 0 {
				values[i] = 0
			}
		}
		b.Labels, b.Values = labels, values
		b.Unit = trim(b.Unit, 12)
		return b, true
	case "divider":
		return b, true
	}
	return Block{}, false
}

func normalizeSide(s Side) (Side, bool) {
	s.Title = trim(s.Title, 40)
	items := make([]string, 0, len(s.Items))
	for _, item := range s.Items {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
		if len(items) == 6 {
			break
		}
	}
	if len(items) == 0 {
		return Side{}, false
	}
	s.Items = items
	return s, true
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// Flatten turns a canvas into compact text for the model's history. The
// renderer keeps the structured payload; this is only what the next turn
// needs to remember.
func Flatten(p Payload) string {
	var b strings.Builder
	if p.Title != "" {
		b.WriteString("【画布】")
		b.WriteString(p.Title)
		b.WriteString("\n")
	}
	for _, block := range p.Blocks {
		switch block.Type {
		case "text":
			b.WriteString(block.Text + "\n")
		case "callout":
			fmt.Fprintf(&b, "[%s] %s %s\n", block.Tone, block.Title, block.Body)
		case "quote":
			fmt.Fprintf(&b, "> %s", block.Text)
			if block.Caption != "" {
				fmt.Fprintf(&b, "（%s）", block.Caption)
			}
			b.WriteString("\n")
		case "steps":
			for i, item := range block.Items {
				fmt.Fprintf(&b, "%d. %s", i+1, item.Title)
				if item.Body != "" {
					fmt.Fprintf(&b, "：%s", item.Body)
				}
				b.WriteString("\n")
			}
		case "table":
			b.WriteString(strings.Join(block.Columns, " | ") + "\n")
			for _, row := range block.Rows {
				b.WriteString(strings.Join(row, " | ") + "\n")
			}
		case "compare":
			if block.Left != nil {
				fmt.Fprintf(&b, "%s：%s\n", block.Left.Title, strings.Join(block.Left.Items, "；"))
			}
			if block.Right != nil {
				fmt.Fprintf(&b, "%s：%s\n", block.Right.Title, strings.Join(block.Right.Items, "；"))
			}
		case "code":
			fmt.Fprintf(&b, "代码（%s）：\n%s\n", block.Language, block.Code)
		case "flashcards":
			for _, card := range block.Cards {
				fmt.Fprintf(&b, "记忆卡：%s → %s\n", card.Front, card.Back)
			}
		case "bar":
			for i, label := range block.Labels {
				fmt.Fprintf(&b, "%s: %g%s\n", label, block.Values[i], block.Unit)
			}
		case "divider":
			// no content
		}
	}
	out := strings.TrimSpace(b.String())
	r := []rune(out)
	if len(r) > 2500 {
		return string(r[:2500]) + "…"
	}
	return out
}

// Types returns the normalized block types, for the tool result.
func Types(p Payload) string {
	out := make([]string, 0, len(p.Blocks))
	for _, b := range p.Blocks {
		out = append(out, b.Type)
	}
	return strings.Join(out, ",")
}
