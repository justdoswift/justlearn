package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/agent"
	"github.com/justdoswift/justlearn/desktop/internal/store"
)

// TestAgentLive runs one real tutor turn. It needs PI_API_KEY in ../.env or
// the environment and is skipped otherwise.
func TestAgentLive(t *testing.T) {
	loadDotEnv()
	if os.Getenv("PI_API_KEY") == "" || os.Getenv("JUSTLEARN_LIVE") != "1" {
		t.Skip("set JUSTLEARN_LIVE=1 and PI_API_KEY to run the live tutor test")
	}
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "live.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	convs, _ := st.Conversations()
	conv := convs[0]
	engine := &agent.Engine{Store: st, Config: agent.Config{
		BaseURL: os.Getenv("PI_BASE_URL"),
		Model:   os.Getenv("PI_MODEL"),
		APIKey:  os.Getenv("PI_API_KEY"),
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	send := func(input string) (string, []string) {
		var text string
		var tools []string
		err := engine.Send(ctx, conv, input, func(ev agent.Event) {
			switch ev.Kind {
			case "delta":
				text += ev.Text
			case "tool":
				tools = append(tools, ev.Text)
			}
		})
		if err != nil {
			t.Fatalf("send %q: %v (text=%q tools=%v)", input, err, text, tools)
		}
		t.Logf("turn %q: tools=%v chars=%d text=%q", input, tools, len([]rune(text)), text)
		if len([]rune(text)) > 800 {
			t.Fatalf("tutor message is too long for a guided turn: %d chars", len([]rune(text)))
		}
		return text, tools
	}

	// The tutor may ask one diagnostic question before creating the route.
	for _, input := range []string{
		"我想零基础学 SQL，帮我定一条 4 周的学习路线。",
		"我是开发，会一点 SQL，想用在数据分析上。",
		"就按数据分析这个方向，先给我一条 4 周路线。",
	} {
		send(input)
		nodes, _ := st.Nodes(conv.GoalID)
		if len(nodes) >= 2 {
			break
		}
	}
	nodes, _ := st.Nodes(conv.GoalID)
	goal, _ := st.Goal(conv.GoalID)
	if goal.Title == "新的学习路线" || len(nodes) < 2 {
		t.Fatalf("route not created: goal=%q nodes=%d", goal.Title, len(nodes))
	}

	// Start the first node; allow one diagnostic answer before the heatmap.
	var current store.Node
	for _, input := range []string{
		"从第一节开始。",
		"我大概知道 SELECT 和 WHERE，窗口函数没怎么用过。直接开始吧。",
	} {
		send(input)
		current, _ = st.CurrentNode(conv.GoalID)
		if len(current.Cells) > 0 {
			break
		}
	}
	if len(current.Cells) < 3 {
		t.Fatalf("heatmap not defined: %d cells", len(current.Cells))
	}
}
