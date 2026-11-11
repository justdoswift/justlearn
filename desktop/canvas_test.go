package main

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/justdoswift/justlearn/desktop/internal/genui"
	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

func canvasMessage(id string, p genui.Payload) store.Message {
	data, _ := json.Marshal(p)
	return store.Message{ID: id, Role: "assistant", Kind: "ui", Text: p.Title, Payload: string(data)}
}

func demoCanvas() genui.Payload {
	return genui.Payload{
		Title: "主键与索引：分别回答什么？",
		Blocks: []genui.Block{
			{Type: "quote", Text: "它们经常一起出现，但解决的是不同问题。", Caption: "主键与索引的分工"},
			{Type: "compare",
				Left:  &genui.Side{Title: "主键", Items: []string{"唯一标识一行", "不能为 NULL", "一张表一个"}},
				Right: &genui.Side{Title: "索引", Items: []string{"加速定位", "可以有多个", "不改变数据"}},
			},
			{Type: "callout", Tone: "info", Title: "一句话", Body: "主键是身份，索引是目录。"},
			{Type: "table", Columns: []string{"问题", "主键", "索引"}, Rows: [][]string{
				{"回答什么", "这一行是谁", "怎么更快找到"},
				{"会改数据吗", "不会", "不会"},
				{"能有几个", "一个", "多个"},
			}},
			{Type: "steps", Title: "查一条记录时", Items: []genui.Item{
				{Title: "先看条件", Body: "WHERE 命中了哪个索引"},
				{Title: "再定位行", Body: "通过主键找到具体记录"},
				{Title: "最后回表", Body: "需要其他列时再读数据页"},
			}},
			{Type: "code", Language: "sql", Caption: "按主键查一行", Code: "SELECT * FROM students WHERE id = 42;"},
		},
	}
}

func demoCanvasCards() genui.Payload {
	return genui.Payload{
		Title: "把这两个词分开记",
		Blocks: []genui.Block{
			{Type: "flashcards", Title: "点一下翻面", Cards: []genui.Card{
				{Front: "主键", Back: "唯一标识一行，不能重复。"},
				{Front: "索引", Back: "加速查找的目录，不改变数据。"},
				{Front: "回表", Back: "通过索引找到主键后，再读完整行。"},
				{Front: "覆盖索引", Back: "查询需要的列都在索引里，不用回表。"},
			}},
			{Type: "bar", Title: "一次查询扫描多少行", Labels: []string{"全表扫描", "索引查找"}, Values: []float64{1200, 1}, Unit: " 行"},
			{Type: "divider"},
			{Type: "callout", Tone: "success", Title: "检查", Body: "如果 WHERE 里没有可用的索引，数据库可能扫描整张表。"},
		},
	}
}

func TestRenderCanvas(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.messages = []store.Message{
		{ID: "user-canvas", Role: "user", Kind: "text", Text: "帮我对比一下主键和索引。"},
		canvasMessage("canvas-demo", demoCanvas()),
	}
	renderSettled(t, a, filepath.Join("..", ".oil-ui", "redesign", "canvas.png"), 1440, 1280)
	light := *a
	light.dark = false
	renderSettled(t, &light, filepath.Join("..", ".oil-ui", "redesign", "canvas-light.png"), 1440, 1280)
}

func TestRenderCanvasCards(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.messages = []store.Message{
		{ID: "user-cards", Role: "user", Kind: "text", Text: "用卡片帮我记一下。"},
		canvasMessage("canvas-cards", demoCanvasCards()),
	}
	renderSettled(t, a, filepath.Join("..", ".oil-ui", "redesign", "canvas-cards.png"), 1440, 1100)
}

func TestFlashcardFlips(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.messages = []store.Message{
		{ID: "user-cards", Role: "user", Kind: "text", Text: "用卡片帮我记一下。"},
		canvasMessage("canvas-cards", demoCanvasCards()),
	}
	tt := ui.NewTester(a.view, 1240, 900)
	tt.Frame()
	if tt.HasText("唯一标识一行，不能重复。") {
		t.Fatal("flashcard back should start hidden")
	}
	if err := tt.Click("记忆卡：主键"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("唯一标识一行，不能重复。") {
		t.Fatal("flashcard did not flip")
	}
}
