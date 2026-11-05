// Package store keeps the app's learning data in a local SQLite database.
package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo/plugins/sqlite"
)

type Store struct{ db *sqlite.DB }

type Goal struct {
	ID           string
	Title        string
	Insight      string
	Level        string
	Total        int
	Done         int
	CurrentTitle string
}

func (g Goal) Percent() int {
	if g.Total == 0 {
		return 0
	}
	return g.Done * 100 / g.Total
}

type Node struct {
	ID         string
	GoalID     string
	Seq        int
	StageTitle string
	Title      string
	Summary    string
	Status     string
	Threshold  int
	Cells      []Cell
}

type Cell struct {
	ID           string
	NodeID       string
	Title        string
	Description  string
	Mastery      int
	Evidence     int
	LastEvidence string
}

type Conversation struct {
	ID        string
	GoalID    string
	Title     string
	Subtitle  string
	UpdatedAt time.Time
}

type Message struct {
	ID             string
	ConversationID string
	Role           string
	Kind           string
	Text           string
	Payload        string
	CreatedAt      time.Time
}

type StageInput struct {
	Title   string
	Summary string
	Lessons []LessonInput
}

type LessonInput struct {
	Title      string
	Summary    string
	EstMinutes int
	Objectives []string
}

var ErrNotFound = errors.New("store: not found")

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]))
}

func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sqlite.Open(context.Background(), path, sqlite.OpenOptions{})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	ctx := context.Background()
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		_, _ = db.Execute(ctx, pragma)
	}
	for _, stmt := range schema {
		if _, err := db.Execute(ctx, stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("schema: %w", err)
		}
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

var schema = []string{
	`CREATE TABLE IF NOT EXISTS goals (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		level TEXT NOT NULL DEFAULT '',
		insight TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		goal_id TEXT NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
		seq INTEGER NOT NULL,
		stage INTEGER NOT NULL DEFAULT 1,
		stage_title TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		summary TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending',
		heatmap_threshold INTEGER NOT NULL DEFAULT 2
	)`,
	`CREATE INDEX IF NOT EXISTS nodes_goal_idx ON nodes(goal_id, seq)`,
	`CREATE TABLE IF NOT EXISTS cells (
		id TEXT PRIMARY KEY,
		node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		mastery INTEGER NOT NULL DEFAULT 0,
		evidence INTEGER NOT NULL DEFAULT 0,
		last_evidence TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS cells_node_idx ON cells(node_id)`,
	`CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		goal_id TEXT REFERENCES goals(id) ON DELETE SET NULL,
		title TEXT NOT NULL,
		subtitle TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		kind TEXT NOT NULL DEFAULT 'text',
		text TEXT NOT NULL DEFAULT '',
		payload TEXT NOT NULL DEFAULT '{}',
		created_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS messages_conv_idx ON messages(conversation_id, created_at)`,
	`CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`,
}

func nowText() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func asInt(v any) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asTime(v any) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, asString(v))
	return t
}

func (s *Store) query(ctx context.Context, sql string, args ...any) (sqlite.Rows, error) {
	return s.db.Query(ctx, sql, args...)
}

// EnsureSeed creates the first empty route and conversation on a fresh install.
func (s *Store) EnsureSeed() error {
	ctx := context.Background()
	rows, err := s.query(ctx, `SELECT COUNT(*) FROM conversations`)
	if err != nil {
		return err
	}
	if len(rows.Values) > 0 && asInt(rows.Values[0][0]) > 0 {
		return nil
	}
	now := nowText()
	goalID := newID()
	convID := newID()
	if _, err := s.db.Execute(ctx, `INSERT INTO goals (id,title,level,insight,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		goalID, "新的学习路线", "", "", now, now); err != nil {
		return err
	}
	if _, err := s.db.Execute(ctx, `INSERT INTO conversations (id,goal_id,title,subtitle,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		convID, goalID, "新对话", "开始新的学习", now, now); err != nil {
		return err
	}
	return nil
}

func (s *Store) Goals() ([]Goal, error) {
	rows, err := s.query(context.Background(), `
		SELECT g.id, g.title, g.insight, g.level,
		       COUNT(n.id) AS total,
		       COALESCE(SUM(CASE WHEN n.status = 'done' THEN 1 ELSE 0 END), 0) AS done,
		       COALESCE((SELECT title FROM nodes WHERE goal_id = g.id AND status = 'current' ORDER BY seq LIMIT 1), '') AS current
		FROM goals g LEFT JOIN nodes n ON n.goal_id = g.id
		GROUP BY g.id ORDER BY g.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	var out []Goal
	for _, row := range rows.Values {
		g := Goal{ID: asString(row[0]), Title: asString(row[1]), Insight: asString(row[2]), Level: asString(row[3]),
			Total: asInt(row[4]), Done: asInt(row[5]), CurrentTitle: asString(row[6])}
		out = append(out, g)
	}
	return out, nil
}

func (s *Store) Goal(id string) (Goal, error) {
	goals, err := s.Goals()
	if err != nil {
		return Goal{}, err
	}
	for _, g := range goals {
		if g.ID == id {
			return g, nil
		}
	}
	return Goal{}, ErrNotFound
}

func (s *Store) Conversations() ([]Conversation, error) {
	rows, err := s.query(context.Background(), `
		SELECT c.id, COALESCE(c.goal_id,''), c.title, c.subtitle, c.updated_at
		FROM conversations c ORDER BY c.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	var out []Conversation
	for _, row := range rows.Values {
		out = append(out, Conversation{
			ID: asString(row[0]), GoalID: asString(row[1]), Title: asString(row[2]),
			Subtitle: asString(row[3]), UpdatedAt: asTime(row[4]),
		})
	}
	return out, nil
}

func (s *Store) Conversation(id string) (Conversation, error) {
	rows, err := s.query(context.Background(), `
		SELECT id, COALESCE(goal_id,''), title, subtitle, updated_at FROM conversations WHERE id = ?`, id)
	if err != nil || len(rows.Values) == 0 {
		return Conversation{}, ErrNotFound
	}
	row := rows.Values[0]
	return Conversation{ID: asString(row[0]), GoalID: asString(row[1]), Title: asString(row[2]),
		Subtitle: asString(row[3]), UpdatedAt: asTime(row[4])}, nil
}

func (s *Store) CreateRoute(title string) (Goal, Conversation, error) {
	ctx := context.Background()
	now := nowText()
	goalID, convID := newID(), newID()
	if title == "" {
		title = "新的学习路线"
	}
	if _, err := s.db.Execute(ctx, `INSERT INTO goals (id,title,level,insight,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		goalID, title, "", "", now, now); err != nil {
		return Goal{}, Conversation{}, err
	}
	if _, err := s.db.Execute(ctx, `INSERT INTO conversations (id,goal_id,title,subtitle,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		convID, goalID, "新对话", "开始新的学习", now, now); err != nil {
		return Goal{}, Conversation{}, err
	}
	return Goal{ID: goalID, Title: title}, Conversation{ID: convID, GoalID: goalID, Title: "新对话", Subtitle: "开始新的学习", UpdatedAt: time.Now()}, nil
}

func (s *Store) CreateConversation(goalID string) (Conversation, error) {
	ctx := context.Background()
	now := nowText()
	id := newID()
	if _, err := s.db.Execute(ctx, `INSERT INTO conversations (id,goal_id,title,subtitle,created_at,updated_at) VALUES (?,?,?,?,?,?)`,
		id, nullIfEmpty(goalID), "新对话", "开始新的学习", now, now); err != nil {
		return Conversation{}, err
	}
	return Conversation{ID: id, GoalID: goalID, Title: "新对话", Subtitle: "开始新的学习", UpdatedAt: time.Now()}, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *Store) TouchConversation(id string) error {
	_, err := s.db.Execute(context.Background(), `UPDATE conversations SET updated_at = ? WHERE id = ?`, nowText(), id)
	return err
}

func (s *Store) RenameConversation(id, title, subtitle string) error {
	_, err := s.db.Execute(context.Background(),
		`UPDATE conversations SET title = ?, subtitle = ?, updated_at = ? WHERE id = ?`, title, subtitle, nowText(), id)
	return err
}

func (s *Store) Messages(conversationID string) ([]Message, error) {
	rows, err := s.query(context.Background(), `
		SELECT id, conversation_id, role, kind, text, payload, created_at
		FROM messages WHERE conversation_id = ? ORDER BY created_at, rowid`, conversationID)
	if err != nil {
		return nil, err
	}
	var out []Message
	for _, row := range rows.Values {
		out = append(out, Message{
			ID: asString(row[0]), ConversationID: asString(row[1]), Role: asString(row[2]),
			Kind: asString(row[3]), Text: asString(row[4]), Payload: asString(row[5]), CreatedAt: asTime(row[6]),
		})
	}
	return out, nil
}

func (s *Store) AppendMessage(m *Message) error {
	if m.ID == "" {
		m.ID = newID()
	}
	if m.Kind == "" {
		m.Kind = "text"
	}
	if m.Payload == "" {
		m.Payload = "{}"
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	_, err := s.db.Execute(context.Background(), `
		INSERT INTO messages (id, conversation_id, role, kind, text, payload, created_at)
		VALUES (?,?,?,?,?,?,?)`,
		m.ID, m.ConversationID, m.Role, m.Kind, m.Text, m.Payload, m.CreatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) UpdateMessageText(id, text string) error {
	_, err := s.db.Execute(context.Background(), `UPDATE messages SET text = ? WHERE id = ?`, text, id)
	return err
}

func (s *Store) Nodes(goalID string) ([]Node, error) {
	ctx := context.Background()
	rows, err := s.query(ctx, `
		SELECT id, goal_id, seq, stage_title, title, summary, status, heatmap_threshold
		FROM nodes WHERE goal_id = ? ORDER BY seq`, goalID)
	if err != nil {
		return nil, err
	}
	nodes := make([]Node, 0, len(rows.Values))
	for _, row := range rows.Values {
		nodes = append(nodes, Node{
			ID: asString(row[0]), GoalID: asString(row[1]), Seq: asInt(row[2]),
			StageTitle: asString(row[3]), Title: asString(row[4]), Summary: asString(row[5]),
			Status: asString(row[6]), Threshold: asInt(row[7]),
		})
	}
	cellRows, err := s.query(ctx, `
		SELECT c.id, c.node_id, c.title, c.description, c.mastery, c.evidence, c.last_evidence
		FROM cells c JOIN nodes n ON n.id = c.node_id WHERE n.goal_id = ? ORDER BY c.created_at, c.rowid`, goalID)
	if err != nil {
		return nil, err
	}
	byNode := make(map[string][]Cell)
	for _, row := range cellRows.Values {
		c := Cell{ID: asString(row[0]), NodeID: asString(row[1]), Title: asString(row[2]),
			Description: asString(row[3]), Mastery: asInt(row[4]), Evidence: asInt(row[5]), LastEvidence: asString(row[6])}
		byNode[c.NodeID] = append(byNode[c.NodeID], c)
	}
	for i := range nodes {
		nodes[i].Cells = byNode[nodes[i].ID]
	}
	return nodes, nil
}

func (s *Store) CurrentNode(goalID string) (Node, error) {
	nodes, err := s.Nodes(goalID)
	if err != nil {
		return Node{}, err
	}
	for _, n := range nodes {
		if n.Status == "current" {
			return n, nil
		}
	}
	for _, n := range nodes {
		if n.Status == "pending" {
			return n, nil
		}
	}
	return Node{}, ErrNotFound
}

func (s *Store) SetGoal(goalID, title, level string, stages []StageInput) error {
	ctx := context.Background()
	now := nowText()
	if title == "" {
		title = "新的学习路线"
	}
	if _, err := s.db.Execute(ctx, `UPDATE goals SET title = ?, level = ?, updated_at = ? WHERE id = ?`,
		title, level, now, goalID); err != nil {
		return err
	}
	if len(stages) == 0 {
		rows, err := s.query(ctx, `SELECT COUNT(*) FROM nodes WHERE goal_id = ?`, goalID)
		if err != nil {
			return err
		}
		if len(rows.Values) > 0 && asInt(rows.Values[0][0]) > 0 {
			return nil
		}
		stages = []StageInput{{Title: "入门", Summary: "", Lessons: []LessonInput{
			{Title: "从一个问题开始", Summary: "用具体问题建立这个主题的第一块认知。"},
			{Title: "核心概念", Summary: "理解这个主题最基本的概念。"},
			{Title: "动手练习", Summary: "在真实例子里用一次。"},
		}}}
	}

	existing, err := s.Nodes(goalID)
	if err != nil {
		return err
	}
	byTitle := make(map[string]Node, len(existing))
	for _, node := range existing {
		byTitle[node.Title] = node
	}

	kept := map[string]bool{}
	seq := 0
	sawCurrent := false
	for si, stage := range stages {
		for _, lesson := range stage.Lessons {
			seq++
			if old, ok := byTitle[lesson.Title]; ok {
				kept[old.ID] = true
				status := old.Status
				if status == "current" {
					sawCurrent = true
				}
				if _, err := s.db.Execute(ctx, `
					UPDATE nodes SET seq = ?, stage = ?, stage_title = ?, title = ?, summary = ? WHERE id = ?`,
					seq, si+1, stage.Title, lesson.Title, lesson.Summary, old.ID); err != nil {
					return err
				}
				continue
			}
			if _, err := s.db.Execute(ctx, `
				INSERT INTO nodes (id,goal_id,seq,stage,stage_title,title,summary,status,heatmap_threshold)
				VALUES (?,?,?,?,?,?,?,?,?)`,
				newID(), goalID, seq, si+1, stage.Title, lesson.Title, lesson.Summary, "pending", 2); err != nil {
				return err
			}
		}
	}
	for _, old := range existing {
		if !kept[old.ID] {
			if _, err := s.db.Execute(ctx, `DELETE FROM nodes WHERE id = ?`, old.ID); err != nil {
				return err
			}
		}
	}
	if !sawCurrent {
		rows, err := s.query(ctx, `SELECT id FROM nodes WHERE goal_id = ? ORDER BY seq LIMIT 1`, goalID)
		if err != nil {
			return err
		}
		if len(rows.Values) > 0 {
			if _, err := s.db.Execute(ctx, `UPDATE nodes SET status = 'current' WHERE id = ?`, asString(rows.Values[0][0])); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) DefineHeatmap(nodeID string, threshold int, cells []Cell, replace bool) ([]Cell, int, error) {
	ctx := context.Background()
	rows, err := s.query(ctx, `SELECT id FROM cells WHERE node_id = ?`, nodeID)
	if err != nil {
		return nil, 0, err
	}
	existing := len(rows.Values)
	if existing > 0 && !replace {
		now, err := s.Heatmap(nodeID)
		if err != nil {
			return nil, 0, err
		}
		return now.Cells, now.Threshold, nil
	}
	if existing > 0 {
		if _, err := s.db.Execute(ctx, `DELETE FROM cells WHERE node_id = ?`, nodeID); err != nil {
			return nil, 0, err
		}
	}
	if threshold < 1 {
		threshold = 1
	}
	if threshold > 3 {
		threshold = 3
	}
	now := nowText()
	out := make([]Cell, 0, len(cells))
	for _, cell := range cells {
		cell.ID = newID()
		cell.NodeID = nodeID
		if _, err := s.db.Execute(ctx, `
			INSERT INTO cells (id,node_id,title,description,mastery,evidence,last_evidence,created_at)
			VALUES (?,?,?,?,0,0,'',?)`,
			cell.ID, nodeID, cell.Title, cell.Description, now); err != nil {
			return nil, 0, err
		}
		out = append(out, cell)
	}
	if _, err := s.db.Execute(ctx, `UPDATE nodes SET heatmap_threshold = ? WHERE id = ?`, threshold, nodeID); err != nil {
		return nil, 0, err
	}
	return out, threshold, nil
}

type Heatmap struct {
	NodeID    string
	Threshold int
	Cells     []Cell
}

func (s *Store) Heatmap(nodeID string) (Heatmap, error) {
	ctx := context.Background()
	rows, err := s.query(ctx, `SELECT heatmap_threshold FROM nodes WHERE id = ?`, nodeID)
	if err != nil || len(rows.Values) == 0 {
		return Heatmap{}, ErrNotFound
	}
	h := Heatmap{NodeID: nodeID, Threshold: asInt(rows.Values[0][0])}
	cellRows, err := s.query(ctx, `SELECT id,node_id,title,description,mastery,evidence,last_evidence FROM cells WHERE node_id = ? ORDER BY rowid`, nodeID)
	if err != nil {
		return Heatmap{}, err
	}
	for _, row := range cellRows.Values {
		h.Cells = append(h.Cells, Cell{ID: asString(row[0]), NodeID: asString(row[1]), Title: asString(row[2]),
			Description: asString(row[3]), Mastery: asInt(row[4]), Evidence: asInt(row[5]), LastEvidence: asString(row[6])})
	}
	return h, nil
}

func (s *Store) UpdateCell(cellID string, mastery int, evidence string) (Cell, error) {
	ctx := context.Background()
	if mastery < 0 {
		mastery = 0
	}
	if mastery > 3 {
		mastery = 3
	}
	if _, err := s.db.Execute(ctx,
		`UPDATE cells SET mastery = ?, evidence = evidence + 1, last_evidence = ? WHERE id = ?`,
		mastery, evidence, cellID); err != nil {
		return Cell{}, err
	}
	rows, err := s.query(ctx, `SELECT id,node_id,title,description,mastery,evidence,last_evidence FROM cells WHERE id = ?`, cellID)
	if err != nil || len(rows.Values) == 0 {
		return Cell{}, ErrNotFound
	}
	row := rows.Values[0]
	return Cell{ID: asString(row[0]), NodeID: asString(row[1]), Title: asString(row[2]), Description: asString(row[3]),
		Mastery: asInt(row[4]), Evidence: asInt(row[5]), LastEvidence: asString(row[6])}, nil
}

func (s *Store) AdvanceNode(goalID string) (Node, Node, error) {
	ctx := context.Background()
	nodes, err := s.Nodes(goalID)
	if err != nil {
		return Node{}, Node{}, err
	}
	var current Node
	found := false
	for _, n := range nodes {
		if n.Status == "current" {
			current, found = n, true
			break
		}
	}
	if !found {
		return Node{}, Node{}, errors.New("当前没有可前进的节点")
	}
	if len(current.Cells) == 0 {
		return Node{}, Node{}, errors.New("这一节还没有定义热力图")
	}
	for _, cell := range current.Cells {
		if cell.Mastery < current.Threshold {
			return Node{}, Node{}, fmt.Errorf("还有格子没到阈值 %d/3", current.Threshold)
		}
	}
	if _, err := s.db.Execute(ctx, `UPDATE nodes SET status = 'done' WHERE id = ?`, current.ID); err != nil {
		return Node{}, Node{}, err
	}
	var next Node
	for _, n := range nodes {
		if n.Seq == current.Seq+1 {
			next = n
			if _, err := s.db.Execute(ctx, `UPDATE nodes SET status = 'current' WHERE id = ?`, n.ID); err != nil {
				return Node{}, Node{}, err
			}
			break
		}
	}
	current.Status = "done"
	return current, next, nil
}

func (s *Store) GetSetting(key, fallback string) string {
	rows, err := s.query(context.Background(), `SELECT value FROM settings WHERE key = ?`, key)
	if err != nil || len(rows.Values) == 0 {
		return fallback
	}
	return asString(rows.Values[0][0])
}

func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Execute(context.Background(),
		`INSERT INTO settings (key,value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value)
	return err
}

// HasNodeCard reports whether a conversation already holds a card of kind for
// the node, so the tutor does not save the same lesson twice.
func (s *Store) HasNodeCard(conversationID, kind, nodeID string) (bool, error) {
	rows, err := s.query(context.Background(),
		`SELECT COUNT(*) FROM messages WHERE conversation_id = ? AND kind = ? AND payload LIKE ?`,
		conversationID, kind, "%\"nodeId\":\""+nodeID+"\"%")
	if err != nil {
		return false, err
	}
	return len(rows.Values) > 0 && asInt(rows.Values[0][0]) > 0, nil
}

// DeleteRoute removes a route, its nodes and cells, and its conversations and
// messages.
func (s *Store) DeleteRoute(goalID string) error {
	ctx := context.Background()
	if _, err := s.db.Execute(ctx, `DELETE FROM conversations WHERE goal_id = ?`, goalID); err != nil {
		return err
	}
	_, err := s.db.Execute(ctx, `DELETE FROM goals WHERE id = ?`, goalID)
	return err
}

// DeleteConversation removes a conversation and its messages, leaving its
// route untouched.
func (s *Store) DeleteConversation(id string) error {
	_, err := s.db.Execute(context.Background(), `DELETE FROM conversations WHERE id = ?`, id)
	return err
}
