package store

import (
	"path/filepath"
	"testing"
)

func TestSetGoalKeepsProgress(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "merge.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	convs, _ := st.Conversations()
	goalID := convs[0].GoalID
	initial := []StageInput{{Title: "基础", Lessons: []LessonInput{
		{Title: "A"}, {Title: "B"}, {Title: "C"},
	}}}
	if err := st.SetGoal(goalID, "SQL", "", initial); err != nil {
		t.Fatal(err)
	}
	nodes, _ := st.Nodes(goalID)
	if len(nodes) != 3 {
		t.Fatalf("nodes %d", len(nodes))
	}
	firstID := nodes[0].ID
	if _, _, err := st.DefineHeatmap(firstID, 2, []Cell{{Title: "x"}, {Title: "y"}, {Title: "z"}}, false); err != nil {
		t.Fatal(err)
	}
	nodes, _ = st.Nodes(goalID)
	if _, err := st.UpdateCell(nodes[0].Cells[0].ID, 2, "ok"); err != nil {
		t.Fatal(err)
	}

	edited := []StageInput{{Title: "基础", Lessons: []LessonInput{
		{Title: "A"}, {Title: "B-new"}, {Title: "C"}, {Title: "D"},
	}}}
	if err := st.SetGoal(goalID, "SQL", "", edited); err != nil {
		t.Fatal(err)
	}
	nodes, _ = st.Nodes(goalID)
	if len(nodes) != 4 {
		t.Fatalf("nodes after edit %d", len(nodes))
	}
	if nodes[0].ID != firstID {
		t.Fatalf("first node id changed: %s -> %s", firstID, nodes[0].ID)
	}
	if len(nodes[0].Cells) != 3 || nodes[0].Cells[0].Mastery != 2 {
		t.Fatalf("progress lost: %+v", nodes[0].Cells)
	}
	current, err := st.CurrentNode(goalID)
	if err != nil || current.ID != firstID {
		t.Fatalf("current %v %v", current.ID, err)
	}
}

func TestDeleteRouteAndConversation(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "delete.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	convs, _ := st.Conversations()
	goalID := convs[0].GoalID
	if err := st.SetGoal(goalID, "SQL", "", []StageInput{{Title: "基础", Lessons: []LessonInput{{Title: "A"}, {Title: "B"}}}}); err != nil {
		t.Fatal(err)
	}
	nodes, _ := st.Nodes(goalID)
	if _, _, err := st.DefineHeatmap(nodes[0].ID, 2, []Cell{{Title: "x"}, {Title: "y"}, {Title: "z"}}, false); err != nil {
		t.Fatal(err)
	}
	convID := convs[0].ID
	if err := st.AppendMessage(&Message{ConversationID: convID, Role: "assistant", Kind: "text", Text: "hi"}); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteRoute(goalID); err != nil {
		t.Fatal(err)
	}
	goals, _ := st.Goals()
	if len(goals) != 0 {
		t.Fatalf("goals left: %v", goals)
	}
	if convs, _ := st.Conversations(); len(convs) != 0 {
		t.Fatalf("conversations left: %v", convs)
	}
	if msgs, _ := st.Messages(convID); len(msgs) != 0 {
		t.Fatalf("messages left: %v", msgs)
	}

	// Conversation deletion keeps the route.
	_, conv, err := st.CreateRoute("SQL 2")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.AppendMessage(&Message{ConversationID: conv.ID, Role: "user", Kind: "text", Text: "hello"}); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteConversation(conv.ID); err != nil {
		t.Fatal(err)
	}
	if convs, _ := st.Conversations(); len(convs) != 0 {
		t.Fatalf("conversation not deleted: %v", convs)
	}
}
