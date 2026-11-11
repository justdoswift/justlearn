package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/genui"
	"github.com/justdoswift/justlearn/desktop/internal/store"
)

// Engine runs one tutor turn: it streams the model's text, runs the learning
// tools and stores the messages the app reads.
type Engine struct {
	Store  *store.Store
	Config Config
	mu     sync.Mutex
}

type Event struct {
	Kind    string // delta, message, tool, done
	Text    string
	Message *store.Message
}

func (e *Engine) Send(ctx context.Context, conv store.Conversation, userText string, emit func(Event)) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	cfg := e.Config.withDefaults()

	history, err := e.Store.Messages(conv.ID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(userText) != "" {
		user := &store.Message{ConversationID: conv.ID, Role: "user", Kind: "text", Text: strings.TrimSpace(userText)}
		if err := e.Store.AppendMessage(user); err != nil {
			return err
		}
		emit(Event{Kind: "message", Message: user})
	}

	messages := []chatMessage{{Role: "system", Content: systemPrompt}}
	messages = append(messages, historyMessages(history)...)
	if strings.TrimSpace(userText) != "" {
		messages = append(messages, chatMessage{Role: "user", Content: strings.TrimSpace(userText)})
	}

	tools := Tools()
	toolbox := &toolset{store: e.Store, goalID: conv.GoalID, convID: conv.ID}
	var allText strings.Builder
	var pending strings.Builder
	lastEmit := time.Now()
	flush := func(force bool) {
		if pending.Len() == 0 {
			return
		}
		if !force && time.Since(lastEmit) < 45*time.Millisecond {
			return
		}
		emit(Event{Kind: "delta", Text: pending.String()})
		pending.Reset()
		lastEmit = time.Now()
	}

	for round := 0; round < 10; round++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, err := streamChat(ctx, cfg, conv.ID, messages, tools, func(delta string) {
			allText.WriteString(delta)
			pending.WriteString(delta)
			flush(false)
		})
		flush(true)
		if err != nil {
			return err
		}
		if len(result.ToolCalls) == 0 {
			text := strings.TrimSpace(allText.String())
			if text != "" {
				msg := &store.Message{ConversationID: conv.ID, Role: "assistant", Kind: "text", Text: text}
				if err := e.Store.AppendMessage(msg); err != nil {
					return err
				}
				emit(Event{Kind: "message", Message: msg})
			}
			_ = e.Store.TouchConversation(conv.ID)
			emit(Event{Kind: "done"})
			return nil
		}

		assistant := chatMessage{Role: "assistant", Content: result.Content}
		for _, call := range result.ToolCalls {
			assistant.ToolCalls = append(assistant.ToolCalls, call)
		}
		messages = append(messages, assistant)
		for _, call := range result.ToolCalls {
			output := toolbox.execute(ctx, call.Function.Name, call.Function.Arguments)
			messages = append(messages, chatMessage{Role: "tool", ToolCallID: call.ID, Content: output})
			emit(Event{Kind: "tool", Text: call.Function.Name})
		}
	}
	return fmt.Errorf("模型连续调用工具次数过多，已停止这一轮")
}

func historyMessages(history []store.Message) []chatMessage {
	start := 0
	if len(history) > 40 {
		start = len(history) - 40
	}
	out := make([]chatMessage, 0, len(history)-start)
	for _, msg := range history[start:] {
		switch msg.Kind {
		case "lesson":
			var payload struct {
				Kicker string `json:"kicker"`
				Body   string `json:"body"`
			}
			_ = json.Unmarshal([]byte(msg.Payload), &payload)
			out = append(out, chatMessage{Role: "assistant", Content: "【微课】" + payload.Kicker + "\n" + payload.Body})
		case "quiz":
			var payload struct {
				Questions []struct {
					Prompt  string   `json:"prompt"`
					Options []string `json:"options"`
				} `json:"questions"`
			}
			_ = json.Unmarshal([]byte(msg.Payload), &payload)
			var b strings.Builder
			b.WriteString("【小测】\n")
			for i, q := range payload.Questions {
				fmt.Fprintf(&b, "%d. %s\n", i+1, q.Prompt)
				for j, opt := range q.Options {
					fmt.Fprintf(&b, "   %c. %s\n", 'A'+j, opt)
				}
			}
			out = append(out, chatMessage{Role: "assistant", Content: b.String()})
		case "ui":
			var payload genui.Payload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err == nil {
				if text := genui.Flatten(payload); text != "" {
					out = append(out, chatMessage{Role: "assistant", Content: text})
				}
			}
		default:
			if msg.Role == "user" || msg.Role == "assistant" {
				if strings.TrimSpace(msg.Text) != "" {
					out = append(out, chatMessage{Role: msg.Role, Content: msg.Text})
				}
			}
		}
	}
	return out
}
