package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

func settle(tt *ui.Tester) { tt.Advance(600 * time.Millisecond) }

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
	abilities, err := st.UpsertAbilities(conv.GoalID, []store.AbilityInput{
		{Title: "数据建模", Description: "能把业务对象整理成表、字段和关系。"},
		{Title: "查询思维", Description: "能把一个问题拆成可验证的查询步骤。"},
		{Title: "SQL 操作", Description: "能写出正确、可读、可维护的 SQL。"},
		{Title: "统计判断", Description: "能判断数字是否可信、差异是否有意义。"},
		{Title: "结论表达", Description: "能把分析结果讲成别人能用的结论。"},
	})
	if err != nil {
		t.Fatal(err)
	}
	current, _ := st.CurrentNode(conv.GoalID)
	_, _, err = st.DefineHeatmap(current.ID, 2, []store.Cell{
		{Title: "主键与外键", AbilityID: abilities[0].ID},
		{Title: "字段类型", AbilityID: abilities[0].ID},
		{Title: "表之间的关系", AbilityID: abilities[0].ID},
		{Title: "索引的作用", AbilityID: abilities[2].ID},
		{Title: "范式", AbilityID: abilities[1].ID},
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
	a.current, _ = st.CurrentNode(conv.GoalID)
	nodes, _ = st.Nodes(conv.GoalID)
	a.nodes = nodes
	return st, a
}

func TestHomeStartsCentered(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	if !a.home {
		t.Fatal("app should start on the home state")
	}
	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("centered composer missing")
	}
	want := float32((1240 - readingWidth) / 2)
	if field.X < want-1 || field.X > want+1 {
		t.Fatalf("composer x=%v, want %v", field.X, want)
	}
	if field.Y < 180 || field.Y > 560 {
		t.Fatalf("composer is not vertically centered: y=%v", field.Y)
	}
	if field.H > 60 {
		t.Fatalf("home composer is too tall: h=%v", field.H)
	}
	if _, ok := tt.Find("发送"); ok {
		t.Fatal("empty home should not expose a send button")
	}
	if tt.HasText("开始一段新的学习。你现在最想学会什么？") {
		t.Fatal("home state should not render the chat transcript")
	}
	for _, unwanted := range []string{
		"一次只走一小步。路线和知识热力图会在对话里慢慢长出来。",
		"转行数据分析", "通过日语 N2", "读懂财务报表",
	} {
		if tt.HasText(unwanted) {
			t.Fatalf("home state still contains copy %q", unwanted)
		}
	}
	for _, name := range []string{"设置"} {
		if _, ok := tt.Find(name); !ok {
			t.Fatalf("top-right control %q missing", name)
		}
	}
	if _, ok := tt.Find("路线图"); ok {
		t.Fatal("the home page should not have a route panel button")
	}
}

func TestContinueLastMovesComposerDown(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	homeField, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("home composer missing")
	}
	if err := tt.Click("继续上次 →"); err != nil {
		t.Fatal(err)
	}
	// The same glass field flies to the bottom; it must not jump there.
	tt.Advance(80 * time.Millisecond)
	mid, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("composer missing during the transition")
	}
	if mid.Y <= homeField.Y+10 || mid.Y >= 700 {
		t.Fatalf("composer did not animate to the bottom: home=%v mid=%v", homeField.Y, mid.Y)
	}
	settle(tt)
	if a.home {
		t.Fatal("continue should enter the chat state")
	}
	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("chat composer missing")
	}
	if field.Y < 620 {
		t.Fatalf("chat composer did not move to the bottom: y=%v", field.Y)
	}
	if !tt.HasText("开始一段新的学习。你现在最想学会什么？") {
		t.Fatal("chat transcript missing")
	}
}

func TestComposerAlignsWithReadingColumn(t *testing.T) {
	_, a := seed(t)
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	msg, ok := tt.Find("开始一段新的学习。你现在最想学会什么？")
	if !ok {
		t.Fatal("first message missing")
	}
	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("composer panel missing")
	}
	if diff := msg.X - field.X; diff > 1 || diff < -1 {
		t.Fatalf("composer is not aligned with the reading column: message x=%v, composer x=%v", msg.X, field.X)
	}
	if field.W < readingWidth-1 || field.W > readingWidth+1 {
		t.Fatalf("composer width=%v, want %v", field.W, readingWidth)
	}
}

func TestRouteRailShowsStagesAndPreview(t *testing.T) {
	_, a := seed(t)
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	for _, want := range []string{"阶段 01，数据库基础", "阶段 02，窗口函数"} {
		if _, ok := tt.Find(want); !ok {
			t.Fatalf("route rail missing %q", want)
		}
	}
	if _, ok := tt.Find("路线图"); ok {
		t.Fatal("route panel button should be gone")
	}

	target, ok := tt.Find("阶段 02，窗口函数")
	if !ok {
		t.Fatal("second stage trigger missing")
	}
	tt.Move(target.X+target.W/2, target.Y+target.H/2)
	tt.Frame()
	tt.Advance(80 * time.Millisecond)
	if a.railShown <= 0.01 || a.railShown >= 0.99 {
		t.Fatalf("rail preview did not animate: shown=%v", a.railShown)
	}
	settle(tt)
	if a.railShown < 0.99 {
		t.Fatalf("rail preview did not settle: shown=%v", a.railShown)
	}
	for _, want := range []string{"窗口帧", "排名与移动平均", "阶段 02"} {
		if !tt.HasText(want) {
			t.Fatalf("hovered stage preview missing %q", want)
		}
	}
}

func TestRenderRouteRail(t *testing.T) {
	_, a := seed(t)
	a.home = false
	dir := filepath.Join("..", ".oil-ui", "redesign")
	renderSettled(t, a, filepath.Join(dir, "route-rail.png"), 1440, 900)
	light := *a
	light.dark = false
	renderSettled(t, &light, filepath.Join(dir, "route-rail-light.png"), 1440, 900)
	renderRailHover(t, a, filepath.Join(dir, "route-rail-hover.png"), "阶段 02，窗口函数")
	renderRailHover(t, a, filepath.Join(dir, "route-rail-current.png"), "阶段 01，数据库基础")
	renderRailHover(t, &light, filepath.Join(dir, "route-rail-hover-light.png"), "阶段 02，窗口函数")
}

func renderRailHover(t *testing.T, a *app, path, label string) {
	t.Helper()
	tt := ui.NewTester(a.view, 1440, 900)
	tt.SetScale(2)
	tt.Frame()
	target, ok := tt.Find(label)
	if !ok {
		t.Fatalf("rail trigger %q missing", label)
	}
	tt.Move(target.X+target.W/2, target.Y+target.H/2)
	tt.Frame()
	tt.Advance(700 * time.Millisecond)
	savePNG(t, path, tt.Image())
}

func TestTopRightButtonsOpenPanels(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	tc := struct {
		button string
		title  string
	}{"设置", "设置"}
	if err := tt.Click(tc.button); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if !a.panelOpen {
		t.Fatalf("%s panel did not open", tc.title)
	}
	if !tt.HasText(tc.title) {
		t.Fatalf("%s panel title missing", tc.title)
	}
	if err := tt.Click("关闭"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if a.panelOpen {
		t.Fatalf("%s panel did not close", tc.title)
	}
}

func TestModelMenuStaysInReadingColumn(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.modelMenuOpen = true
	a.modelsLoaded = true
	a.models = []string{"deepseek-v4.1-flash"}
	tt := ui.NewTester(a.view, 1440, 900)
	tt.Frame()
	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("composer missing")
	}
	menu, ok := tt.Find("模型")
	if !ok {
		t.Fatal("model menu missing")
	}
	if menu.X+menu.W > field.X+field.W+1 {
		t.Fatalf("model menu overflows the reading column: menu right=%v composer right=%v",
			menu.X+menu.W, field.X+field.W)
	}
}

func TestHeatPanelFixedRightFades(t *testing.T) {
	_, a := seed(t)
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("composer missing")
	}
	if field.W != readingWidth {
		t.Fatalf("heat panel reflowed the composer: width=%v", field.W)
	}
	if a.heatShown < 0.68 || a.heatShown > 0.76 {
		t.Fatalf("heat panel default fade = %v, want about 0.72", a.heatShown)
	}
	ability, ok := tt.Find("数据建模")
	if !ok {
		t.Fatal("ability row missing")
	}
	if ability.X < 1000 {
		t.Fatalf("ability heatmap is not fixed on the right: row x=%v", ability.X)
	}

	tt.Move(ability.X+ability.W/2, ability.Y+ability.H/2)
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	if a.heatShown < 0.99 {
		t.Fatalf("heat panel did not restore on hover: shown=%v", a.heatShown)
	}
	if a.heatHover != 0 || a.heatCardShown < 0.99 {
		t.Fatalf("heat detail did not open: hover=%d card=%v", a.heatHover, a.heatCardShown)
	}

	tt.Move(420, 420)
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	if a.heatShown > 0.76 {
		t.Fatalf("heat panel did not fade back after hover: shown=%v", a.heatShown)
	}
	if a.heatHover != -1 || a.heatCardShown > 0.01 {
		t.Fatalf("heat detail did not close: hover=%d card=%v", a.heatHover, a.heatCardShown)
	}
}

func TestAbilityNamesLiveInHoverCard(t *testing.T) {
	_, a := seed(t)
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()

	ability, ok := tt.Find("数据建模")
	if !ok {
		t.Fatal("ability row missing")
	}
	if ability.X < 1000 {
		t.Fatalf("ability heatmap is not fixed on the right: row=%v", ability)
	}
	if ability.W > 60 {
		t.Fatalf("ability row still reserves space for a visible label: %v", ability)
	}
	if a.heatCardShown > 0.01 {
		t.Fatalf("ability detail should stay hidden at rest: %v", a.heatCardShown)
	}

	tt.Move(ability.X+ability.W/2, ability.Y+ability.H/2)
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	if a.heatHover != 0 || a.heatCardShown < 0.99 {
		t.Fatalf("heat detail did not open: hover=%d card=%v", a.heatHover, a.heatCardShown)
	}
	for _, want := range []string{"数据建模", "能把业务对象整理成表、字段和关系。", "熟练", "67%", "1 初学", "1 熟练", "1 精通", "覆盖 1 节 · 3 个格子"} {
		if !tt.HasText(want) {
			t.Fatalf("ability detail missing %q", want)
		}
	}
	tip, ok := tt.Find("能力详情")
	if !ok {
		t.Fatal("ability detail card missing")
	}
	if tip.X+tip.W > ability.X+1 {
		t.Fatalf("ability detail card overlaps the rail: tip right=%v row left=%v", tip.X+tip.W, ability.X)
	}
}

func TestHeatHoverMovesWithoutFlicker(t *testing.T) {
	_, a := seed(t)
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	first, ok := tt.Find("数据建模")
	if !ok {
		t.Fatal("first ability row missing")
	}
	second, ok := tt.Find("查询思维")
	if !ok {
		t.Fatal("second ability row missing")
	}

	tt.Move(first.X+first.W/2, first.Y+first.H/2)
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	card1, ok := tt.Find("能力详情")
	if !ok {
		t.Fatal("ability detail card missing")
	}
	if a.heatHover != 0 || a.heatCardShown < 0.99 {
		t.Fatalf("first ability did not open the detail: hover=%d card=%v", a.heatHover, a.heatCardShown)
	}

	tt.Move(second.X+second.W/2, second.Y+second.H/2)
	tt.Frame()
	tt.Advance(50 * time.Millisecond)
	card2, ok := tt.Find("能力详情")
	if !ok {
		t.Fatal("ability detail card disappeared between rows")
	}
	if a.heatHover != 1 {
		t.Fatalf("hover did not move to the second ability: %d", a.heatHover)
	}
	if a.heatCardShown < 0.99 {
		t.Fatalf("ability detail flickered between rows: %v", a.heatCardShown)
	}
	if card1 != card2 {
		t.Fatalf("ability detail moved between rows: %v then %v", card1, card2)
	}
}

func TestAbilityRowsDoNotOverlap(t *testing.T) {
	_, a := seed(t)
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	var last ui.Rect
	for _, ability := range a.abilities {
		row, ok := tt.Find(ability.Title)
		if !ok {
			t.Fatalf("ability row %q missing", ability.Title)
		}
		if last.H > 0 && row.Y < last.Y+last.H {
			t.Fatalf("ability rows overlap: %v then %v", last, row)
		}
		last = row
	}
}

func TestAbilityHeatsAggregateAcrossNodes(t *testing.T) {
	abilities := []store.Ability{{ID: "model", Title: "数据建模"}, {ID: "stats", Title: "统计判断"}}
	nodes := []store.Node{
		{ID: "n1", Cells: []store.Cell{
			{AbilityID: "model", Mastery: 1, Evidence: 1},
			{AbilityID: "stats", Mastery: 2, Evidence: 1},
		}},
		{ID: "n2", Cells: []store.Cell{
			{AbilityID: "model", Mastery: 3, Evidence: 2},
			{AbilityID: "stats", Mastery: 0, Evidence: 1},
		}},
	}
	heats := abilityHeats(abilities, nodes)
	if heats[0].Nodes != 2 || len(heats[0].Cells) != 2 || heats[0].Average != 2 || heats[0].Mastered != 1 {
		t.Fatalf("model heat = %+v, want 2 nodes, 2 cells, average 2, 1 mastered", heats[0])
	}
	if heats[1].Nodes != 2 || len(heats[1].Cells) != 2 || heats[1].Average != 1 {
		t.Fatalf("stats heat = %+v, want 2 nodes, 2 cells, average 1", heats[1])
	}
}

func TestRenderHeatmap(t *testing.T) {
	_, a := seed(t)
	decorateHeat(t, a)
	a.home = false
	dir := filepath.Join("..", ".oil-ui", "redesign")
	renderSettled(t, a, filepath.Join(dir, "heatmap.png"), 1440, 900)
	light := *a
	light.dark = false
	renderSettled(t, &light, filepath.Join(dir, "heatmap-light.png"), 1440, 900)
	renderHeatTip(t, a, filepath.Join(dir, "heatmap-tip.png"), "数据建模")
	renderHeatTip(t, &light, filepath.Join(dir, "heatmap-tip-light.png"), "数据建模")
}

func renderHeatTip(t *testing.T, a *app, path, label string) {
	t.Helper()
	tt := ui.NewTester(a.view, 1440, 900)
	tt.SetScale(2)
	tt.Frame()
	tile, ok := tt.Find(label)
	if !ok {
		t.Fatalf("heat row %q missing", label)
	}
	if tile.X < 1000 {
		t.Fatalf("heat row %q is not in the right rail: %v", label, tile)
	}
	tt.Move(tile.X+tile.W/2, tile.Y+tile.H/2)
	tt.Frame()
	tt.Advance(700 * time.Millisecond)
	savePNG(t, path, tt.Image())
}

// decorateHeat gives later nodes sample cells so the render shows the full
// neutral-to-jade scale. Tests never assert on these extra cells.
func decorateHeat(t *testing.T, a *app) {
	t.Helper()
	nodes, err := a.store.Nodes(a.goal.ID)
	if err != nil {
		t.Fatal(err)
	}
	labels := []string{"概念边界", "判断标准", "典型例子", "常见误区", "迁移应用"}
	for ni, node := range nodes {
		if len(node.Cells) > 0 {
			continue
		}
		cells := make([]store.Cell, 0, 5)
		for _, label := range labels {
			abilityID := ""
			if len(a.abilities) > 0 {
				abilityID = a.abilities[len(cells)%len(a.abilities)].ID
			}
			cells = append(cells, store.Cell{Title: label, Description: node.Title + "：" + label, AbilityID: abilityID})
		}
		inserted, _, err := a.store.DefineHeatmap(node.ID, 2, cells, false)
		if err != nil {
			t.Fatal(err)
		}
		for ci, cell := range inserted {
			mastery := (ni + ci) % 4
			switch {
			case ni == 1:
				mastery = 3
			case ni == 2:
				mastery = 2
			case ni == 3:
				mastery = 1
			case ni == len(nodes)-1:
				mastery = 0
			}
			if _, err := a.store.UpdateCell(cell.ID, mastery, "渲染样本"); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := a.reloadForConversation(a.active); err != nil {
		t.Fatal(err)
	}
}

func TestPanelAnimates(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(80 * time.Millisecond)
	if a.panelShown <= 0.01 || a.panelShown >= 0.99 {
		t.Fatalf("panel did not animate: shown=%v", a.panelShown)
	}
	settle(tt)
	if a.panelShown < 0.99 {
		t.Fatalf("panel did not settle: shown=%v", a.panelShown)
	}
}

func TestAppearanceSetting(t *testing.T) {
	_, a := seed(t)
	a.openSettings()
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	settle(tt)
	if !a.settingsOpen {
		t.Fatal("settings did not open")
	}
	if err := tt.Click("浅色"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if a.dark || a.appearance != "light" {
		t.Fatalf("light appearance not applied: dark=%v appearance=%q", a.dark, a.appearance)
	}
	if got := a.store.GetSetting("appearance", ""); got != "light" {
		t.Fatalf("light appearance not saved: %q", got)
	}
	if err := tt.Click("深色"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if !a.dark || a.appearance != "dark" {
		t.Fatalf("dark appearance not applied: dark=%v appearance=%q", a.dark, a.appearance)
	}
	if err := tt.Click("跟随系统"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if a.appearance != "system" {
		t.Fatalf("system appearance not applied: %q", a.appearance)
	}
}

func TestComposerFitsMinimumWindow(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1120, 580)
	tt.Frame()
	for _, name := range []string{"输入区", "导入文本", "发送", "DeepSeek V4.1 Flash"} {
		rect, ok := tt.Find(name)
		if !ok {
			continue
		}
		if rect.Y < 0 || rect.Y+rect.H > 580 {
			t.Fatalf("%s is outside the window: %+v", name, rect)
		}
	}
}
