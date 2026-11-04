package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"desktop/internal/agent"
	"desktop/internal/store"
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
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	var text string
	var tools []string
	err = engine.Send(ctx, conv, "我想零基础学 SQL，帮我定一条 4 周的学习路线。", func(ev agent.Event) {
		switch ev.Kind {
		case "delta":
			text += ev.Text
		case "tool":
			tools = append(tools, ev.Text)
		}
	})
	if err != nil {
		t.Fatalf("send: %v (text=%q tools=%v)", err, text, tools)
	}
	nodes, _ := st.Nodes(conv.GoalID)
	goal, _ := st.Goal(conv.GoalID)
	t.Logf("goal=%q nodes=%d tools=%v text=%q", goal.Title, len(nodes), tools, text)
	if goal.Title == "新的学习路线" || len(nodes) < 2 {
		t.Fatalf("route not created: goal=%q nodes=%d", goal.Title, len(nodes))
	}

	// Second turn: the tutor should define the heatmap and teach the first node.
	text = ""
	tools = nil
	err = engine.Send(ctx, conv, "从第一节开始。", func(ev agent.Event) {
		switch ev.Kind {
		case "delta":
			text += ev.Text
		case "tool":
			tools = append(tools, ev.Text)
		}
	})
	if err != nil {
		t.Fatalf("lesson send: %v (text=%q tools=%v)", err, text, tools)
	}
	current, _ := st.CurrentNode(conv.GoalID)
	msgs, _ := st.Messages(conv.ID)
	kinds := map[string]int{}
	for _, m := range msgs {
		kinds[m.Kind]++
	}
	t.Logf("lesson turn: tools=%v cells=%d threshold=%d kinds=%v", tools, len(current.Cells), current.Threshold, kinds)
	if len(current.Cells) < 3 {
		t.Fatalf("heatmap not defined: %d cells", len(current.Cells))
	}
	if kinds["lesson"] == 0 || kinds["quiz"] == 0 {
		t.Fatalf("lesson/quiz missing: %v", kinds)
	}

	// Third turn: answer the first quiz question and let the tutor score it.
	var answer string
	for _, m := range msgs {
		if m.Kind != "quiz" {
			continue
		}
		var qp quizPayload
		_ = json.Unmarshal([]byte(m.Payload), &qp)
		if len(qp.Questions) > 0 && len(qp.Questions[0].Options) > 0 {
			opt := qp.Questions[0].Options[0]
			answer = fmt.Sprintf("第 1 题（%s）我选 A：%s。用我自己的话说，选这个是因为它唯一且不随业务变化。", qp.Questions[0].Prompt, opt)
		}
	}
	if answer == "" {
		t.Fatal("no quiz question to answer")
	}
	text = ""
	tools = nil
	err = engine.Send(ctx, conv, answer, func(ev agent.Event) {
		switch ev.Kind {
		case "delta":
			text += ev.Text
		case "tool":
			tools = append(tools, ev.Text)
		}
	})
	if err != nil {
		t.Fatalf("answer send: %v", err)
	}
	current, _ = st.CurrentNode(conv.GoalID)
	scored := 0
	evidence := 0
	for _, cell := range current.Cells {
		if cell.Mastery > 0 {
			scored++
		}
		if cell.Evidence > 0 {
			evidence++
		}
	}
	t.Logf("answer turn: tools=%v scored=%d/%d evidence=%d text=%q", tools, scored, len(current.Cells), evidence, text)
	if evidence == 0 {
		t.Fatalf("heatmap not updated after answer: %+v", current.Cells)
	}
}
