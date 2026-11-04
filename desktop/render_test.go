package main

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

func seed(t *testing.T) (*store.Store, *app) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	_ = st.SetSetting("model", "deepseek-v4.1-flash")
	convs, _ := st.Conversations()
	conv := convs[0]
	if err := st.SetGoal(conv.GoalID, "3 个月掌握数据分析", "零基础", []store.StageInput{
		{Title: "数据库基础", Lessons: []store.LessonInput{
			{Title: "数据库与表结构", Summary: "表和字段"},
			{Title: "SELECT 与过滤", Summary: "查询"},
			{Title: "聚合与分组", Summary: "GROUP BY"},
		}},
		{Title: "窗口函数", Lessons: []store.LessonInput{
			{Title: "窗口帧", Summary: "OVER"},
			{Title: "排名与移动平均", Summary: "RANK"},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	current, _ := st.CurrentNode(conv.GoalID)
	_, _, err = st.DefineHeatmap(current.ID, 2, []store.Cell{
		{Title: "主键与外键"}, {Title: "字段类型"}, {Title: "表之间的关系"},
		{Title: "索引的作用"}, {Title: "范式"},
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	nodes, _ := st.Nodes(conv.GoalID)
	for i, mastery := range []int{1, 2, 3, 2, 1} {
		_, err := st.UpdateCell(nodes[0].Cells[i].ID, mastery, "测试证据")
		if err != nil {
			t.Fatal(err)
		}
	}
	_ = st.AppendMessage(&store.Message{ConversationID: conv.ID, Role: "assistant", Kind: "text",
		Text: "开始一段新的学习。你现在最想学会什么？"})
	_ = st.AppendMessage(&store.Message{ConversationID: conv.ID, Role: "user", Kind: "text",
		Text: "我想转行做数据分析，零基础。"})
	lesson := lessonPayload{
		NodeID: nodes[0].ID, Kicker: "第 1 节 · 数据库与表结构",
		Title: "一张表到底长什么样",
		Body:  "数据库不是一堆文件，而是**一组有约束的表格**。\n\n- 每一行是一条记录\n- 每一列是一个字段，有自己的类型\n\n## 一个例子\n\n学生表里，`id` 是主键，`class_id` 指向班级表的主键。",
		Recall: struct {
			Prompt string `json:"prompt"`
			Answer string `json:"answer"`
		}{Prompt: "为什么 id 适合做主键？", Answer: "唯一、稳定、不随业务变化。"},
		KeyPoints: []string{"行是记录，列是字段", "主键唯一标识一行"},
	}
	payload, _ := json.Marshal(lesson)
	_ = st.AppendMessage(&store.Message{ConversationID: conv.ID, Role: "assistant", Kind: "lesson",
		Text: lesson.Title, Payload: string(payload)})
	quiz := quizPayload{NodeID: nodes[0].ID, Questions: []struct {
		ID          string   `json:"id"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		AnswerIndex int      `json:"answerIndex"`
		Explanation string   `json:"explanation"`
	}{{ID: "q1", Prompt: "选出可以作为主键的字段", Options: []string{"学号", "姓名", "年龄"}, AnswerIndex: 0, Explanation: "学号唯一且稳定。"}}}
	qdata, _ := json.Marshal(quiz)
	_ = st.AppendMessage(&store.Message{ConversationID: conv.ID, Role: "assistant", Kind: "quiz",
		Text: "小测", Payload: string(qdata)})

	a := newApp(st)
	a.active = conv
	if err := a.reloadForConversation(conv); err != nil {
		t.Fatal(err)
	}
	a.routeOpen = true
	a.current = current
	nodes, _ = st.Nodes(conv.GoalID)
	a.nodes = nodes
	return st, a
}

func TestRender(t *testing.T) {
	_, a := seed(t)
	img := ui.Render(a.view, 1240, 820, 2)
	f, err := os.Create("testdata/ui.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestRenderEmpty(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	a := newApp(st)
	a.routeOpen = true
	img := ui.Render(a.view, 1440, 900, 2)
	f, err := os.Create("testdata/empty.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestInteractions(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("模型设置") {
		t.Fatalf("settings dialog missing, texts=%q", tt.Texts())
	}
	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.settingsOpen {
		t.Fatal("settings stayed open")
	}
	if err := tt.Click("新对话"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.messages) != 0 {
		t.Fatalf("new conversation has %d messages", len(a.messages))
	}
	if !tt.HasText("开始一段新的学习") {
		t.Fatalf("empty state missing: %q", tt.Texts())
	}
	// Route panel switch and node selection.
	if err := tt.Click("SELECT 与过滤"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.selectedNode().Title != "SELECT 与过滤" {
		t.Fatalf("selected %q", a.selectedNode().Title)
	}
}
