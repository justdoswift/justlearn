package genui

import (
	"strings"
	"testing"
)

func TestNormalizeKeepsSupportedBlocks(t *testing.T) {
	p := Payload{Title: "主键与索引", Blocks: []Block{
		{Type: "text", Text: "先看区别。"},
		{Type: "callout", Tone: "warning", Body: "索引不是主键。"},
		{Type: "steps", Items: []Item{{Title: "找行"}, {Title: "回表"}}},
		{Type: "table", Columns: []string{"项目", "主键"}, Rows: [][]string{{"作用", "身份"}, {"唯一"}}},
		{Type: "compare", Left: &Side{Title: "主键", Items: []string{"唯一"}}, Right: &Side{Title: "索引", Items: []string{"加速"}}},
		{Type: "code", Language: "sql", Code: "select 1;"},
		{Type: "flashcards", Cards: []Card{{Front: "主键", Back: "身份"}, {Front: "索引", Back: "目录"}}},
		{Type: "bar", Labels: []string{"扫描", "命中"}, Values: []float64{100, 3}, Unit: " 行"},
		{Type: "divider"},
		{Type: "unknown", Text: "drop"},
	}}
	got, err := Normalize(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Blocks) != MaxBlocks {
		t.Fatalf("blocks=%d, want %d", len(got.Blocks), MaxBlocks)
	}
	if got.Blocks[3].Rows[1][1] != "" {
		t.Fatalf("short table row not padded: %#v", got.Blocks[3].Rows)
	}
	if got.Blocks[1].Tone != "warning" {
		t.Fatalf("tone=%q", got.Blocks[1].Tone)
	}
}

func TestNormalizeQuote(t *testing.T) {
	got, err := Normalize(Payload{Blocks: []Block{
		{Type: "quote", Text: "先预测，再核对。", Caption: "阅读方法"},
		{Type: "quote", Text: "   "},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Blocks) != 1 || got.Blocks[0].Caption != "阅读方法" {
		t.Fatalf("quote not normalized: %#v", got.Blocks)
	}
	if flat := Flatten(got); !strings.Contains(flat, "> 先预测，再核对。（阅读方法）") {
		t.Fatalf("Flatten() missing quote: %s", flat)
	}
}

func TestNormalizeCardsBarsAndDivider(t *testing.T) {
	p := Payload{Blocks: []Block{
		{Type: "flashcards", Cards: []Card{{Front: "问", Back: "答"}, {Front: "问题", Back: "答案"}}},
		{Type: "bar", Labels: []string{"A", "B"}, Values: []float64{2, 1}, Unit: "次"},
		{Type: "divider"},
	}}
	got, err := Normalize(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Blocks) != 3 {
		t.Fatalf("blocks=%d, want 3", len(got.Blocks))
	}
	if got.Blocks[2].Type != "divider" {
		t.Fatalf("divider dropped: %#v", got.Blocks)
	}
}

func TestNormalizeRejectsEmptyCanvas(t *testing.T) {
	if _, err := Normalize(Payload{Blocks: []Block{{Type: "text"}, {Type: "table", Columns: []string{"a"}}}}); err == nil {
		t.Fatal("empty canvas should be rejected")
	}
}

func TestFlattenKeepsStructure(t *testing.T) {
	p := Payload{Title: "对比", Blocks: []Block{
		{Type: "table", Columns: []string{"A", "B"}, Rows: [][]string{{"1", "2"}}},
		{Type: "flashcards", Cards: []Card{{Front: "问", Back: "答"}}},
	}}
	got := Flatten(p)
	for _, want := range []string{"【画布】对比", "A | B", "1 | 2", "问 → 答"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Flatten() missing %q: %s", want, got)
		}
	}
}

func TestNormalizeCapsRowsAndColumns(t *testing.T) {
	p := Payload{Blocks: []Block{{
		Type:    "table",
		Columns: []string{"1", "2", "3", "4", "5", "6", "7"},
		Rows:    make([][]string, 20),
	}}}
	got, err := Normalize(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Blocks[0].Columns) != MaxCols || len(got.Blocks[0].Rows) != MaxRows {
		t.Fatalf("table not capped: cols=%d rows=%d", len(got.Blocks[0].Columns), len(got.Blocks[0].Rows))
	}
}
