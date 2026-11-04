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
