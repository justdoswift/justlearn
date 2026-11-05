package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/justdoswift/justlearn/desktop/internal/store"
)

type toolset struct {
	store       *store.Store
	goalID      string
	convID      string
	savedLesson bool
	savedQuiz   bool
}

func schema(properties map[string]any, required ...string) map[string]any {
	m := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		m["required"] = required
	}
	return m
}

func obj(props map[string]any) map[string]any { return schema(props) }

func arr(items map[string]any, min, max int) map[string]any {
	m := map[string]any{"type": "array", "items": items}
	if min > 0 {
		m["minItems"] = min
	}
	if max > 0 {
		m["maxItems"] = max
	}
	return m
}

func str(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func num(desc string, min, max int) map[string]any {
	return map[string]any{"type": "integer", "description": desc, "minimum": min, "maximum": max}
}

func Tools() []Tool {
	return []Tool{
		{Type: "function", Function: ToolFunction{
			Name:        "learning_state",
			Description: "读取当前目标、当前节点和热力图格子（含格子 ID）。每轮开始时先调用。",
			Parameters:  schema(map[string]any{}),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "set_goal",
			Description: "保存或更新学习目标和路线。没有路线时按目标生成 2–6 个阶段、每阶段 1–10 个节点。",
			Parameters: schema(map[string]any{
				"title": str("目标标题"),
				"level": str("当前基础水平"),
				"stages": arr(obj(map[string]any{
					"title":   str("阶段标题"),
					"summary": str("阶段说明"),
					"lessons": arr(obj(map[string]any{
						"title":   str("节点标题，一个概念"),
						"summary": str("节点说明"),
					}), 1, 10),
				}), 2, 6),
			}, "title", "stages"),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "define_heatmap",
			Description: "为一个节点定义 3–9 个知识格子和掌握阈值（1–3）。开始教这一节前调用；用户要求改格子或阈值时也调用，replace=true 会重写格子。",
			Parameters: schema(map[string]any{
				"nodeId":    str("节点 ID；省略则用当前节点"),
				"threshold": num("掌握阈值", 1, 3),
				"cells": arr(obj(map[string]any{
					"title":       str("格子标题，一个可判断的知识侧面"),
					"description": str("这个格子具体测什么"),
				}), 3, 9),
				"replace": map[string]any{"type": "boolean", "description": "重写已有格子"},
			}, "threshold", "cells"),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "update_heatmap",
			Description: "测试完一个知识点后，把某个格子的掌握度更新到 0–3，并写一句证据。只能根据用户真实回答更新。",
			Parameters: schema(map[string]any{
				"cellId":   str("格子 ID，来自 learning_state 或 define_heatmap"),
				"mastery":  num("掌握度：0 没学会，1 初学，2 熟练，3 精通", 0, 3),
				"evidence": str("一句证据，引用用户怎么回答的"),
			}, "cellId", "mastery"),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "advance_node",
			Description: "当前节点的所有格子都达到阈值时调用，把节点标记为掌握并前进到下一节。",
			Parameters:  schema(map[string]any{}),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "save_lesson",
			Description: "只在用户已经通过几轮问答弄懂一个侧面后，保存一张 200–300 字的要点卡；只总结一个点，不要写成整篇文章。",
			Parameters: schema(map[string]any{
				"title": str("微课标题"),
				"body":  str("Markdown 正文"),
				"recall": obj(map[string]any{
					"prompt": str("主动回忆的问题"),
					"answer": str("参考答案"),
				}),
				"code":      str("可选示例代码"),
				"mnemonic":  str("一句记忆钩子"),
				"keyPoints": arr(str("要点"), 0, 6),
			}, "title", "body", "recall"),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "save_quiz",
			Description: "一次只保存 1 道小测题；等用户回答并更新热力图后，下一轮再出下一题。options 至少 2 个，answerIndex 从 0 开始。",
			Parameters: schema(map[string]any{
				"questions": arr(obj(map[string]any{
					"prompt":      str("题目"),
					"options":     arr(str("选项"), 2, 4),
					"answerIndex": num("正确选项下标，从 0 开始", 0, 3),
					"explanation": str("一句解释"),
				}), 1, 1),
			}, "questions"),
		}},
	}
}

func (t *toolset) currentNode() (store.Node, error) {
	return t.store.CurrentNode(t.goalID)
}

func cellList(cells []store.Cell) string {
	if len(cells) == 0 {
		return "未定义"
	}
	parts := make([]string, 0, len(cells))
	for _, cell := range cells {
		parts = append(parts, fmt.Sprintf("id=%s|%s:%d/3", cell.ID, cell.Title, cell.Mastery))
	}
	return strings.Join(parts, "; ")
}

func (t *toolset) execute(ctx context.Context, name, raw string) string {
	switch name {
	case "learning_state":
		goal, err := t.store.Goal(t.goalID)
		if err != nil {
			return "读取失败：" + err.Error()
		}
		nodes, err := t.store.Nodes(t.goalID)
		if err != nil {
			return "读取失败：" + err.Error()
		}
		current, err := t.currentNode()
		if err != nil {
			return fmt.Sprintf("[tool:learning_state] 目标=%s; 节点数=%d; 当前节点=无（路线已完成或还没生成）", goal.Title, len(nodes))
		}
		return fmt.Sprintf("[tool:learning_state] 目标=%s; 进度=%d%%; 节点数=%d; 当前节点=id=%s|第 %d 节 %s; 阈值=%d; 热力图=%s",
			goal.Title, goal.Percent(), len(nodes), current.ID, current.Seq, current.Title, current.Threshold, cellList(current.Cells))

	case "set_goal":
		var args struct {
			Title  string `json:"title"`
			Level  string `json:"level"`
			Stages []struct {
				Title   string `json:"title"`
				Summary string `json:"summary"`
				Lessons []struct {
					Title   string `json:"title"`
					Summary string `json:"summary"`
				} `json:"lessons"`
			} `json:"stages"`
		}
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		if len(args.Stages) < 2 || len(args.Stages) > 6 {
			return "阶段数量要在 2–6 之间。"
		}
		stages := make([]store.StageInput, 0, len(args.Stages))
		for _, stage := range args.Stages {
			lessons := make([]store.LessonInput, 0, len(stage.Lessons))
			for _, lesson := range stage.Lessons {
				if strings.TrimSpace(lesson.Title) == "" {
					continue
				}
				lessons = append(lessons, store.LessonInput{Title: lesson.Title, Summary: lesson.Summary})
			}
			if len(lessons) == 0 {
				continue
			}
			stages = append(stages, store.StageInput{Title: stage.Title, Summary: stage.Summary, Lessons: lessons})
		}
		if err := t.store.SetGoal(t.goalID, args.Title, args.Level, stages); err != nil {
			return "保存路线失败：" + err.Error()
		}
		_ = t.store.RenameConversation(t.convID, args.Title, "学习路线")
		nodes, _ := t.store.Nodes(t.goalID)
		return fmt.Sprintf("[tool:set_goal] ok; title=%s; nodes=%d", args.Title, len(nodes))

	case "define_heatmap":
		var args struct {
			NodeID    string `json:"nodeId"`
			Threshold int    `json:"threshold"`
			Cells     []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"cells"`
			Replace bool `json:"replace"`
		}
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		if len(args.Cells) < 3 || len(args.Cells) > 9 {
			return "格子数量要在 3–9 之间。"
		}
		if args.Threshold < 1 || args.Threshold > 3 {
			return "阈值要在 1–3 之间。"
		}
		node, err := t.currentNode()
		if args.NodeID != "" {
			nodes, _ := t.store.Nodes(t.goalID)
			found := false
			for _, n := range nodes {
				if n.ID == args.NodeID {
					node, found = n, true
					break
				}
			}
			if !found {
				return "找不到这个节点。"
			}
		} else if err != nil {
			return "没有可定义的节点。"
		}
		cells := make([]store.Cell, 0, len(args.Cells))
		for _, cell := range args.Cells {
			if strings.TrimSpace(cell.Title) == "" {
				continue
			}
			cells = append(cells, store.Cell{Title: cell.Title, Description: cell.Description})
		}
		inserted, threshold, err := t.store.DefineHeatmap(node.ID, args.Threshold, cells, args.Replace)
		if err != nil {
			return "保存热力图失败：" + err.Error()
		}
		return fmt.Sprintf("[tool:define_heatmap] ok; node=%s; threshold=%d; cells=%s", node.Title, threshold, cellList(inserted))

	case "update_heatmap":
		var args struct {
			CellID   string `json:"cellId"`
			Mastery  int    `json:"mastery"`
			Evidence string `json:"evidence"`
		}
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		cell, err := t.store.UpdateCell(args.CellID, args.Mastery, args.Evidence)
		if err != nil {
			return "找不到这个格子，或它不属于当前路线。"
		}
		return fmt.Sprintf("[tool:update_heatmap] %s=%d/3; evidence=%s", cell.Title, cell.Mastery, args.Evidence)

	case "advance_node":
		done, next, err := t.store.AdvanceNode(t.goalID)
		if err != nil {
			return "不能前进：" + err.Error()
		}
		nextTitle := "无"
		if next.ID != "" {
			nextTitle = next.Title
		}
		return fmt.Sprintf("[tool:advance_node] mastered=%s; next=%s", done.Title, nextTitle)

	case "save_lesson":
		var args struct {
			Title  string `json:"title"`
			Body   string `json:"body"`
			Recall struct {
				Prompt string `json:"prompt"`
				Answer string `json:"answer"`
			} `json:"recall"`
			Code      string   `json:"code"`
			Mnemonic  string   `json:"mnemonic"`
			KeyPoints []string `json:"keyPoints"`
		}
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		if t.savedQuiz {
			return "这一轮已经出过题了：先停下来等用户回答，下一轮再保存要点卡。"
		}
		if len([]rune(args.Body)) > 500 {
			return fmt.Sprintf("正文太长（%d 字）：一张卡只讲一个点，控制在 200–300 字；需要更多内容请在聊天里分多轮讲。", len([]rune(args.Body)))
		}
		if len(args.KeyPoints) > 3 {
			return "要点最多 3 个：一张卡只总结一个侧面。"
		}
		node, err := t.currentNode()
		if err != nil {
			return "当前没有可保存的节点。"
		}
		if exists, _ := t.store.HasNodeCard(t.convID, "lesson", node.ID); exists {
			return "[tool:save_lesson] already saved for this node"
		}
		payload := map[string]any{
			"nodeId":    node.ID,
			"kicker":    fmt.Sprintf("第 %d 节 · %s", node.Seq, node.Title),
			"title":     args.Title,
			"body":      args.Body,
			"recall":    args.Recall,
			"code":      args.Code,
			"mnemonic":  args.Mnemonic,
			"keyPoints": args.KeyPoints,
		}
		data, _ := json.Marshal(payload)
		msg := &store.Message{ConversationID: t.convID, Role: "assistant", Kind: "lesson",
			Text: args.Title, Payload: string(data)}
		if err := t.store.AppendMessage(msg); err != nil {
			return "保存微课失败：" + err.Error()
		}
		t.savedLesson = true
		return fmt.Sprintf("[tool:save_lesson] ok; kicker=第 %d 节 · %s; title=%s", node.Seq, node.Title, args.Title)

	case "save_quiz":
		var args struct {
			Questions []struct {
				Prompt      string   `json:"prompt"`
				Options     []string `json:"options"`
				AnswerIndex int      `json:"answerIndex"`
				Explanation string   `json:"explanation"`
			} `json:"questions"`
		}
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		if t.savedLesson {
			return "这一轮已经给过要点卡了：先停下等用户回答，下一轮再出题。"
		}
		if len(args.Questions) != 1 {
			return "一次只出 1 道题：等用户回答并更新热力图后，下一轮再出下一题。"
		}
		node, err := t.currentNode()
		if err != nil {
			return "当前没有可出题的节点。"
		}
		production := regexp.MustCompile(`(写|填|算|列出|改写|辨认|翻译|构造|给出|选出|判断.{0,6}(该不该|是否))`)
		hasProduction := false
		questions := make([]map[string]any, 0, len(args.Questions))
		for _, q := range args.Questions {
			if len(q.Options) < 2 || len(q.Options) > 4 || q.AnswerIndex < 0 || q.AnswerIndex >= len(q.Options) {
				return "每道题的选项要有 2–4 个，answerIndex 要在选项范围内。"
			}
			if production.MatchString(q.Prompt) {
				hasProduction = true
			}
			questions = append(questions, map[string]any{
				"id":          fmt.Sprintf("%s-q%d", node.ID[:8], len(questions)+1),
				"prompt":      q.Prompt,
				"options":     q.Options,
				"answerIndex": q.AnswerIndex,
				"explanation": q.Explanation,
			})
		}
		if !hasProduction {
			return "至少一道产出型题：让用户实际写出、选出、算出或改写一个具体结果。"
		}
		if exists, _ := t.store.HasNodeCard(t.convID, "quiz", node.ID); exists {
			return "[tool:save_quiz] already saved for this node"
		}
		payload := map[string]any{"nodeId": node.ID, "questions": questions}
		data, _ := json.Marshal(payload)
		msg := &store.Message{ConversationID: t.convID, Role: "assistant", Kind: "quiz",
			Text: "小测", Payload: string(data)}
		if err := t.store.AppendMessage(msg); err != nil {
			return "保存小测失败：" + err.Error()
		}
		t.savedQuiz = true
		return fmt.Sprintf("[tool:save_quiz] ok; questions=%d", len(questions))
	}
	return "未知工具：" + name
}
