package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/justdoswift/justlearn/desktop/internal/genui"
	"github.com/justdoswift/justlearn/desktop/internal/store"
)

type toolset struct {
	store       *store.Store
	goalID      string
	convID      string
	savedLesson bool
	savedQuiz   bool
	savedCanvas bool
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
			Description: "读取当前目标、当前节点、能力和知识格子（含格子 ID）。每轮开始时先调用。",
			Parameters:  schema(map[string]any{}),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "set_goal",
			Description: "保存或更新学习目标和路线。没有路线时按目标生成 2–6 个阶段、每阶段 1–10 个节点；同时定义 4–6 个跨节点的能力维度。",
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
				"abilities": arr(obj(map[string]any{
					"title":       str("能力名，2–6 个汉字，跨节点可迁移"),
					"description": str("这项能力具体指什么"),
				}), 4, 6),
			}, "title", "stages"),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "set_abilities",
			Description: "定义或调整整条路线的 4–6 个能力维度。已有同名能力会保留 ID 和证据；没有能力时先调用这个。",
			Parameters: schema(map[string]any{
				"abilities": arr(obj(map[string]any{
					"title":       str("能力名，2–6 个汉字，跨节点可迁移"),
					"description": str("这项能力具体指什么"),
				}), 4, 6),
			}, "abilities"),
		}},
		{Type: "function", Function: ToolFunction{
			Name:        "define_heatmap",
			Description: "为一个节点定义 3–9 个知识格子，每个格子必须归属一个能力 ID；开始教这一节前调用。replace=true 会重写格子。",
			Parameters: schema(map[string]any{
				"nodeId":    str("节点 ID；省略则用当前节点"),
				"threshold": num("掌握阈值", 1, 3),
				"cells": arr(obj(map[string]any{
					"title":       str("格子标题，一个可判断的知识侧面"),
					"description": str("这个格子具体测什么"),
					"abilityId":   str("该格子服务的能力 ID，来自 learning_state 或 set_goal"),
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
		{Type: "function", Function: ToolFunction{
			Name:        "render_canvas",
			Description: "把这一轮内容渲染成聊天画布上的原生结构化组件。只在对比、步骤、表格、代码、记忆卡或小图表明显优于纯文字时调用；一次最多 1 张画布、6 个 block。不要生成 HTML/CSS/JS，也不要重复聊天正文。",
			Parameters: schema(map[string]any{
				"title": str("画布标题，可选；不要和正文重复"),
				"blocks": arr(obj(map[string]any{
					"type":    map[string]any{"type": "string", "enum": []string{"text", "callout", "quote", "steps", "table", "compare", "code", "flashcards", "bar", "divider"}, "description": "block 类型"},
					"text":    str("text 类型的 Markdown 正文；quote 类型填引用内容"),
					"tone":    str("callout 语气：info / success / warning"),
					"title":   str("block 标题；steps、table、compare、bar 可用"),
					"body":    str("callout 的正文"),
					"caption": str("可选的图表 / 表格 / 代码 / 引用来源说明"),
					"columns": arr(str("表格列名"), 2, 6),
					"rows":    arr(arr(str("单元格"), 0, 6), 1, 12),
					"items": arr(obj(map[string]any{
						"title": str("步骤标题"),
						"body":  str("步骤说明"),
					}), 2, 6),
					"left": obj(map[string]any{
						"title": str("左侧标题"),
						"items": arr(str("左侧条目"), 1, 6),
					}),
					"right": obj(map[string]any{
						"title": str("右侧标题"),
						"items": arr(str("右侧条目"), 1, 6),
					}),
					"language": str("代码语言，例如 sql / go / python"),
					"code":     str("代码正文"),
					"cards": arr(obj(map[string]any{
						"front": str("记忆卡正面"),
						"back":  str("记忆卡背面"),
					}), 2, 8),
					"labels": arr(str("柱状图标签"), 2, 8),
					"values": arr(map[string]any{"type": "number"}, 2, 8),
					"unit":   str("图表单位，例如 行 / 秒 / %"),
				}), 1, 6),
			}, "blocks"),
		}},
	}
}

func (t *toolset) currentNode() (store.Node, error) {
	return t.store.CurrentNode(t.goalID)
}

func abilityNames(abilities []store.Ability) map[string]string {
	names := make(map[string]string, len(abilities))
	for _, ability := range abilities {
		names[ability.ID] = ability.Title
	}
	return names
}

func abilityList(abilities []store.Ability) string {
	if len(abilities) == 0 {
		return "未定义（先调用 set_abilities）"
	}
	parts := make([]string, 0, len(abilities))
	for _, ability := range abilities {
		parts = append(parts, fmt.Sprintf("id=%s|%s", ability.ID, ability.Title))
	}
	return strings.Join(parts, "; ")
}

func cellList(cells []store.Cell, names map[string]string) string {
	if len(cells) == 0 {
		return "未定义"
	}
	parts := make([]string, 0, len(cells))
	for _, cell := range cells {
		ability := names[cell.AbilityID]
		if ability == "" {
			ability = "未归类"
		}
		parts = append(parts, fmt.Sprintf("id=%s|%s|能力=%s:%d/3", cell.ID, cell.Title, ability, cell.Mastery))
	}
	return strings.Join(parts, "; ")
}

func cleanAbilities(input []struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}) ([]store.AbilityInput, error) {
	out := make([]store.AbilityInput, 0, len(input))
	seen := map[string]bool{}
	for _, ability := range input {
		title := strings.TrimSpace(ability.Title)
		if title == "" {
			return nil, fmt.Errorf("能力名不能为空")
		}
		if seen[title] {
			return nil, fmt.Errorf("能力名「%s」重复了", title)
		}
		seen[title] = true
		out = append(out, store.AbilityInput{Title: title, Description: strings.TrimSpace(ability.Description)})
	}
	if len(out) < 4 || len(out) > 6 {
		return nil, fmt.Errorf("能力数量要在 4–6 个之间")
	}
	return out, nil
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
		abilities, err := t.store.Abilities(t.goalID)
		if err != nil {
			return "读取失败：" + err.Error()
		}
		names := abilityNames(abilities)
		current, err := t.currentNode()
		if err != nil {
			return fmt.Sprintf("[tool:learning_state] 目标=%s; 节点数=%d; 当前节点=无（路线已完成或还没生成）; 能力=%s",
				goal.Title, len(nodes), abilityList(abilities))
		}
		return fmt.Sprintf("[tool:learning_state] 目标=%s; 进度=%d%%; 节点数=%d; 当前节点=id=%s|第 %d 节 %s; 阈值=%d; 能力=%s; 当前格子=%s",
			goal.Title, goal.Percent(), len(nodes), current.ID, current.Seq, current.Title, current.Threshold,
			abilityList(abilities), cellList(current.Cells, names))

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
			Abilities []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"abilities"`
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
		existingAbilities, err := t.store.Abilities(t.goalID)
		if err != nil {
			return "读取能力失败：" + err.Error()
		}
		var abilities []store.AbilityInput
		if len(args.Abilities) > 0 {
			abilities, err = cleanAbilities(args.Abilities)
			if err != nil {
				return err.Error()
			}
		} else if len(existingAbilities) == 0 {
			return "先定义 4–6 个能力维度：abilities 不能省略。"
		}
		if err := t.store.SetGoal(t.goalID, args.Title, args.Level, stages); err != nil {
			return "保存路线失败：" + err.Error()
		}
		if len(abilities) > 0 {
			if _, err := t.store.UpsertAbilities(t.goalID, abilities); err != nil {
				return "保存能力失败：" + err.Error()
			}
		}
		_ = t.store.RenameConversation(t.convID, args.Title, "学习路线")
		nodes, _ := t.store.Nodes(t.goalID)
		savedAbilities, _ := t.store.Abilities(t.goalID)
		return fmt.Sprintf("[tool:set_goal] ok; title=%s; nodes=%d; abilities=%d; %s",
			args.Title, len(nodes), len(savedAbilities), abilityList(savedAbilities))

	case "set_abilities":
		var args struct {
			Abilities []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"abilities"`
		}
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		abilities, err := cleanAbilities(args.Abilities)
		if err != nil {
			return err.Error()
		}
		saved, err := t.store.UpsertAbilities(t.goalID, abilities)
		if err != nil {
			return "保存能力失败：" + err.Error()
		}
		return fmt.Sprintf("[tool:set_abilities] ok; %s", abilityList(saved))

	case "define_heatmap":
		var args struct {
			NodeID    string `json:"nodeId"`
			Threshold int    `json:"threshold"`
			Cells     []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				AbilityID   string `json:"abilityId"`
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
		abilities, err := t.store.Abilities(t.goalID)
		if err != nil {
			return "读取能力失败：" + err.Error()
		}
		if len(abilities) == 0 {
			return "还没有能力维度：先调用 set_abilities 定义 4–6 个能力。"
		}
		names := abilityNames(abilities)
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
			ability := strings.TrimSpace(cell.AbilityID)
			if ability == "" {
				return fmt.Sprintf("格子「%s」没有 abilityId：每个格子都要归属一个能力。", cell.Title)
			}
			if names[ability] == "" {
				return fmt.Sprintf("格子「%s」的 abilityId 不属于当前路线；可用：%s", cell.Title, abilityList(abilities))
			}
			cells = append(cells, store.Cell{Title: cell.Title, Description: cell.Description, AbilityID: ability})
		}
		if len(cells) < 3 {
			return "格子数量要在 3–9 之间。"
		}
		inserted, threshold, err := t.store.DefineHeatmap(node.ID, args.Threshold, cells, args.Replace)
		if err != nil {
			return "保存热力图失败：" + err.Error()
		}
		return fmt.Sprintf("[tool:define_heatmap] ok; node=%s; threshold=%d; cells=%s", node.Title, threshold, cellList(inserted, names))

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

	case "render_canvas":
		if t.savedCanvas || t.savedLesson || t.savedQuiz {
			return "这一轮已经给过一张卡或一道题：先停下来等用户回应，下一轮再画布。"
		}
		var args genui.Payload
		if err := json.Unmarshal([]byte(raw), &args); err != nil {
			return "参数解析失败：" + err.Error()
		}
		payload, err := genui.Normalize(args)
		if err != nil {
			return err.Error()
		}
		if node, err := t.currentNode(); err == nil {
			payload.NodeID = node.ID
		}
		data, _ := json.Marshal(payload)
		text := strings.TrimSpace(payload.Title)
		if text == "" {
			text = "画布"
		}
		msg := &store.Message{ConversationID: t.convID, Role: "assistant", Kind: "ui",
			Text: text, Payload: string(data)}
		if err := t.store.AppendMessage(msg); err != nil {
			return "保存画布失败：" + err.Error()
		}
		t.savedCanvas = true
		return fmt.Sprintf("[tool:render_canvas] ok; blocks=%d; types=%s", len(payload.Blocks), genui.Types(payload))

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
		if t.savedQuiz || t.savedCanvas {
			return "这一轮已经给过内容了：先停下来等用户回答，下一轮再保存要点卡。"
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
		if t.savedLesson || t.savedCanvas {
			return "这一轮已经给过内容了：先停下等用户回答，下一轮再出题。"
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
