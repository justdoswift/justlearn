package server

import (
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"
)

// The store's structs carry no JSON tags: they are the desktop app's own
// types, read in Go. The web client gets its own shapes here, in camelCase,
// so the two sides can evolve without one renaming the other's fields.

type goalDTO struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Level        string `json:"level"`
	Done         int    `json:"done"`
	Total        int    `json:"total"`
	Percent      int    `json:"percent"`
	CurrentTitle string `json:"currentTitle"`
}

type conversationDTO struct {
	ID        string    `json:"id"`
	GoalID    string    `json:"goalId"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type cellDTO struct {
	ID           string `json:"id"`
	NodeID       string `json:"nodeId"`
	AbilityID    string `json:"abilityId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Mastery      int    `json:"mastery"`
	Evidence     int    `json:"evidence"`
	LastEvidence string `json:"lastEvidence"`
}

type nodeDTO struct {
	ID         string    `json:"id"`
	GoalID     string    `json:"goalId"`
	Seq        int       `json:"seq"`
	StageTitle string    `json:"stageTitle"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Status     string    `json:"status"`
	Threshold  int       `json:"threshold"`
	Cells      []cellDTO `json:"cells"`
}

type abilityDTO struct {
	ID          string `json:"id"`
	GoalID      string `json:"goalId"`
	Seq         int    `json:"seq"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type messageDTO struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	Role           string    `json:"role"`
	Kind           string    `json:"kind"`
	Text           string    `json:"text"`
	Payload        string    `json:"payload"`
	CreatedAt      time.Time `json:"createdAt"`
}

func toGoal(g store.Goal) goalDTO {
	return goalDTO{
		ID: g.ID, Title: g.Title, Level: g.Level,
		Done: g.Done, Total: g.Total, Percent: g.Percent(),
		CurrentTitle: g.CurrentTitle,
	}
}

func toGoals(goals []store.Goal) []goalDTO {
	out := make([]goalDTO, 0, len(goals))
	for _, g := range goals {
		out = append(out, toGoal(g))
	}
	return out
}

func toConversation(c store.Conversation) conversationDTO {
	return conversationDTO{
		ID: c.ID, GoalID: c.GoalID, Title: c.Title,
		Subtitle: c.Subtitle, UpdatedAt: c.UpdatedAt,
	}
}

func toConversations(convs []store.Conversation) []conversationDTO {
	out := make([]conversationDTO, 0, len(convs))
	for _, c := range convs {
		out = append(out, toConversation(c))
	}
	return out
}

func toCell(c store.Cell) cellDTO {
	return cellDTO{
		ID: c.ID, NodeID: c.NodeID, AbilityID: c.AbilityID,
		Title: c.Title, Description: c.Description,
		Mastery: c.Mastery, Evidence: c.Evidence, LastEvidence: c.LastEvidence,
	}
}

func toNode(n store.Node) nodeDTO {
	cells := make([]cellDTO, 0, len(n.Cells))
	for _, c := range n.Cells {
		cells = append(cells, toCell(c))
	}
	return nodeDTO{
		ID: n.ID, GoalID: n.GoalID, Seq: n.Seq, StageTitle: n.StageTitle,
		Title: n.Title, Summary: n.Summary, Status: n.Status,
		Threshold: n.Threshold, Cells: cells,
	}
}

func toNodes(nodes []store.Node) []nodeDTO {
	out := make([]nodeDTO, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, toNode(n))
	}
	return out
}

func toAbilities(abilities []store.Ability) []abilityDTO {
	out := make([]abilityDTO, 0, len(abilities))
	for _, a := range abilities {
		out = append(out, abilityDTO{
			ID: a.ID, GoalID: a.GoalID, Seq: a.Seq,
			Title: a.Title, Description: a.Description,
		})
	}
	return out
}

func toMessage(m store.Message) messageDTO {
	return messageDTO{
		ID: m.ID, ConversationID: m.ConversationID, Role: m.Role,
		Kind: m.Kind, Text: m.Text, Payload: m.Payload, CreatedAt: m.CreatedAt,
	}
}

func toMessages(msgs []store.Message) []messageDTO {
	out := make([]messageDTO, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, toMessage(m))
	}
	return out
}
