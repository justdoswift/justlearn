package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

const formatSQL = `with order_cnt as (
  select
    user_id,
    category_id,
    count(distinct order_id) as order_cnt
  from orders o
  where o.created_at >= date '2024-05-01'
    and o.created_at < date '2024-06-01'
    and o.status = 'paid'
  group by 1, 2
)
select
  category_id,
  count(*) filter (where order_cnt >= 2) as repeat_buyer_cnt,
  count(*) as buyer_cnt
from user_cat
group by 1;`

func formatMessages(t *testing.T) (store.Message, store.Message) {
	t.Helper()
	lesson := lessonPayload{
		NodeID: "node-format", Kicker: "第 2 节 · 复购率",
		Title: "先把粒度说清楚",
		Body:  "不要一上来就写 `count(*)`。先说清楚这一行代表什么，再决定分母。",
		Recall: struct {
			Prompt string `json:"prompt"`
			Answer string `json:"answer"`
		}{Prompt: "group by 之后一行代表什么？", Answer: "一个「品类 × 用户」粒度。"},
		Code: formatSQL,
		KeyPoints: []string{
			"先写清「一行代表什么」，再写 group by",
			"业务词必须先落成 where / join 条件",
			"分子分母要限定在同一群用户上",
		},
	}
	lessonData, _ := json.Marshal(lesson)
	quiz := quizPayload{NodeID: "node-format", Questions: []struct {
		ID          string   `json:"id"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		AnswerIndex int      `json:"answerIndex"`
		Explanation string   `json:"explanation"`
	}{{
		ID:     "q-format",
		Prompt: "orders 表中一个用户可有多条订单。要回答「上月有多少复购用户（≥2 单）」，哪段写法最终数的是「人」？",
		Options: []string{
			"select count(*) from orders where created_at >= date '2024-05-01' group by user_id",
			"select count(*) from (select user_id from orders where created_at >= date '2024-05-01' and created_at < date '2024-06-01' group by user_id having count(distinct order_id) >= 2) t",
			"select count(user_id) from orders where created_at between date '2024-05-01' and date '2024-05-31' and order_id >= 1",
		},
		AnswerIndex: 1,
		Explanation: "先按 user_id 聚合成一行一个用户，再数这些行。",
	}}}
	quizData, _ := json.Marshal(quiz)
	return store.Message{ID: "lesson-format", Role: "assistant", Kind: "lesson", Text: lesson.Title, Payload: string(lessonData)},
		store.Message{ID: "quiz-format", Role: "assistant", Kind: "quiz", Text: "小测", Payload: string(quizData)}
}

func TestRenderChatFormat(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.dark = false
	lesson, quiz := formatMessages(t)
	a.messages = []store.Message{
		{ID: "user-format", Role: "user", Kind: "text", Text: "帮我算一下复购率。"},
		lesson,
		quiz,
	}
	renderSettled(t, a, filepath.Join("..", ".oil-ui", "redesign", "chat-format.png"), 1440, 1300)
	renderSettled(t, a, filepath.Join("..", ".oil-ui", "redesign", "chat-format-900.png"), 1440, 900)
}

func TestRenderLegacyOptions(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.dark = false
	a.messages = []store.Message{{
		ID:   "legacy-options",
		Role: "assistant", Kind: "text", Text: legacyQuizSample(),
	}}
	renderSettled(t, a, filepath.Join("..", ".oil-ui", "redesign", "legacy-options.png"), 1440, 900)
}

func TestLastMessageStaysAboveComposer(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.messages = append(a.messages, store.Message{
		ID: "last-message", Role: "assistant", Kind: "text",
		Text: "最后一行不应该被输入框压住。",
	})
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	settle(tt)
	msg, ok := tt.Find("最后一行不应该被输入框压住。")
	if !ok {
		t.Fatal("last message missing")
	}
	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("composer missing")
	}
	if msg.Y+msg.H > field.Y-4 {
		t.Fatalf("last message reaches the composer: message bottom=%v composer top=%v", msg.Y+msg.H, field.Y)
	}
}

func TestLongCodeFoldsAndExpands(t *testing.T) {
	_, a := seed(t)
	a.home = false
	lesson, _ := formatMessages(t)
	a.messages = []store.Message{lesson}
	tt := ui.NewTester(a.view, 1240, 900)
	tt.Frame()
	lines := len(strings.Split(formatSQL, "\n"))
	more := fmt.Sprintf("展开全部 %d 行", lines)
	if !tt.HasText(more) {
		t.Fatalf("code block should start folded; looking for %q", more)
	}
	if err := tt.Click(more); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("收起代码") {
		t.Fatal("expanded code block should offer to collapse")
	}
}
