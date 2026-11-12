package agent

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justdoswift/justlearn/desktop/internal/genui"
	"github.com/justdoswift/justlearn/desktop/internal/store"
)

func TestRenderCanvasToolStoresTypedMessage(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "canvas.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	convs, err := st.Conversations()
	if err != nil || len(convs) == 0 {
		t.Fatalf("conversations: %v %d", err, len(convs))
	}
	conv := convs[0]
	tool := &toolset{store: st, goalID: conv.GoalID, convID: conv.ID}
	payload := genui.Payload{
		Title: "主键与索引",
		Blocks: []genui.Block{
			{Type: "text", Text: "先看区别。"},
			{Type: "table", Columns: []string{"对象", "用途"}, Rows: [][]string{{"主键", "身份"}, {"索引", "目录"}}},
		},
	}
	raw, _ := json.Marshal(payload)
	result := tool.execute(context.Background(), "render_canvas", string(raw))
	if !strings.Contains(result, "blocks=2") {
		t.Fatalf("unexpected tool result: %s", result)
	}
	msgs, err := st.Messages(conv.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, msg := range msgs {
		if msg.Kind != "ui" {
			continue
		}
		found = true
		var got genui.Payload
		if err := json.Unmarshal([]byte(msg.Payload), &got); err != nil {
			t.Fatal(err)
		}
		if got.Title != "主键与索引" || len(got.Blocks) != 2 {
			t.Fatalf("stored payload = %#v", got)
		}
	}
	if !found {
		t.Fatal("no ui message stored")
	}
	if again := tool.execute(context.Background(), "render_canvas", string(raw)); !strings.Contains(again, "已经") {
		t.Fatalf("second canvas should be rejected: %s", again)
	}
}

func TestAbilityToolsBindCellsToAbilities(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "abilities.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	convs, err := st.Conversations()
	if err != nil || len(convs) == 0 {
		t.Fatalf("conversations: %v %d", err, len(convs))
	}
	conv := convs[0]
	tool := &toolset{store: st, goalID: conv.GoalID, convID: conv.ID}

	withoutAbilities := `{"title":"数据分析","level":"零基础","stages":[
		{"title":"基础","lessons":[{"title":"表结构"},{"title":"查询"}]},
		{"title":"进阶","lessons":[{"title":"聚合"}]}
	]}`
	if got := tool.execute(context.Background(), "set_goal", withoutAbilities); !strings.Contains(got, "abilities") {
		t.Fatalf("set_goal without abilities should be rejected: %s", got)
	}

	withAbilities := `{"title":"数据分析","level":"零基础","stages":[
		{"title":"基础","lessons":[{"title":"表结构"},{"title":"查询"}]},
		{"title":"进阶","lessons":[{"title":"聚合"}]}
	],"abilities":[
		{"title":"数据建模","description":"能把业务对象整理成表和关系。"},
		{"title":"查询思维","description":"能把问题拆成查询步骤。"},
		{"title":"SQL 操作","description":"能写出正确的 SQL。"},
		{"title":"统计判断","description":"能判断数字是否可信。"},
		{"title":"结论表达","description":"能讲清分析结论。"}
	],"heatmap":{"threshold":2,"cells":[
		{"title":"主键","ability":"数据建模"},
		{"title":"外键","ability":"数据建模"},
		{"title":"索引","ability":"SQL 操作"}
	]}}`
	if got := tool.execute(context.Background(), "set_goal", withAbilities); !strings.Contains(got, "abilities=5") || !strings.Contains(got, "heatmap=3") {
		t.Fatalf("set_goal with abilities failed: %s", got)
	}
	abilities, err := st.Abilities(conv.GoalID)
	if err != nil || len(abilities) != 5 {
		t.Fatalf("abilities: %v %d", err, len(abilities))
	}
	initial, err := st.Nodes(conv.GoalID)
	if err != nil {
		t.Fatal(err)
	}
	if len(initial[0].Cells) != 3 || initial[0].Cells[0].AbilityID != abilities[0].ID {
		t.Fatalf("initial heatmap was not bound to abilities: %+v", initial[0].Cells)
	}

	withoutAbility := `{"threshold":2,"cells":[{"title":"主键"},{"title":"外键"},{"title":"索引"}]}`
	if got := tool.execute(context.Background(), "define_heatmap", withoutAbility); !strings.Contains(got, "abilityId") {
		t.Fatalf("cell without ability should be rejected: %s", got)
	}
	badAbility := `{"threshold":2,"cells":[{"title":"主键","abilityId":"missing"},{"title":"外键","abilityId":"missing"},{"title":"索引","abilityId":"missing"}]}`
	if got := tool.execute(context.Background(), "define_heatmap", badAbility); !strings.Contains(got, "不属于") {
		t.Fatalf("cell with an unknown ability should be rejected: %s", got)
	}

	valid := `{"threshold":2,"replace":true,"cells":[
		{"title":"主键","abilityId":"` + abilities[0].ID + `"},
		{"title":"外键","abilityId":"` + abilities[0].ID + `"},
		{"title":"索引","abilityId":"` + abilities[2].ID + `"}
	]}`
	if got := tool.execute(context.Background(), "define_heatmap", valid); !strings.Contains(got, "能力=数据建模") {
		t.Fatalf("define_heatmap did not report the ability: %s", got)
	}
	nodes, err := st.Nodes(conv.GoalID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes[0].Cells) != 3 || nodes[0].Cells[0].AbilityID != abilities[0].ID {
		t.Fatalf("cells were not bound to abilities: %+v", nodes[0].Cells)
	}
	if got := tool.execute(context.Background(), "learning_state", `{}`); !strings.Contains(got, "能力=") || !strings.Contains(got, "数据建模") {
		t.Fatalf("learning_state did not expose abilities: %s", got)
	}
}

func TestRenderCanvasToolRejectsEmptyPayload(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "canvas.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	convs, _ := st.Conversations()
	tool := &toolset{store: st, goalID: convs[0].GoalID, convID: convs[0].ID}
	result := tool.execute(context.Background(), "render_canvas", `{"blocks":[]}`)
	if !strings.Contains(result, "没有可渲染") {
		t.Fatalf("unexpected result: %s", result)
	}
}
