package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/agent"
	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

// The gear must open the real settings page. It used to open a blank sheet
// with a title and a close button, which read as a broken app.
func TestGearOpensRealSettings(t *testing.T) {
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if !a.settingsOpen {
		t.Fatal("the gear did not open settings")
	}
	for _, want := range []string{"外观", "模型", "关于"} {
		if !tt.HasText(want) {
			t.Fatalf("settings page missing %q", want)
		}
	}
}

// Every route and conversation must be reachable. Before the nav panel the
// only conversation the UI could show was the one it happened to select.
func TestNavReachesEveryConversation(t *testing.T) {
	_, a := seed(t)
	second, err := a.store.CreateConversation(a.goal.ID)
	if err != nil {
		t.Fatal(err)
	}
	_ = a.store.RenameConversation(second.ID, "第二个对话", "")
	if err := a.reloadAll(); err != nil {
		t.Fatal(err)
	}
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	if err := tt.Click("路线与对话"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if !tt.HasText("第二个对话") {
		t.Fatal("the nav panel does not list the second conversation")
	}
	if err := tt.Click("第二个对话"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if a.active.ID != second.ID {
		t.Fatalf("clicking the conversation did not open it: active=%s want=%s", a.active.ID, second.ID)
	}
	if a.navOpen {
		t.Fatal("the nav panel stayed open after choosing a conversation")
	}
}

func TestNavReachesEveryRoute(t *testing.T) {
	_, a := seed(t)
	other, _, err := a.store.CreateRoute("另一条路线")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.reloadAll(); err != nil {
		t.Fatal(err)
	}
	a.home = false
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	if err := tt.Click("路线与对话"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if err := tt.Click("另一条路线"); err != nil {
		t.Fatal(err)
	}
	settle(tt)
	if a.goal.ID != other.ID {
		t.Fatalf("clicking the route did not switch to it: goal=%s want=%s", a.goal.ID, other.ID)
	}
}

// A modal owns the keyboard: Cmd+Enter must not send a draft from under the
// settings sheet or a delete confirmation.
func TestCmdEnterIgnoredWhileModalOpen(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.draft = "不该被发出去"
	a.openSettings()
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyEnter)
	tt.Frame()
	if a.streaming {
		t.Fatal("Cmd+Enter sent a message while settings was open")
	}
	if a.draft != "不该被发出去" {
		t.Fatalf("the draft was consumed: %q", a.draft)
	}
}

// A tool call only changes the route and the ability profile. It must not
// pull a reader who scrolled up back to the end of the transcript.
func TestToolEventKeepsScrollPosition(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.chatScrollToEnd = false
	a.handleEvent(agent.Event{Kind: "tool", Text: "learning_state"})
	if a.chatScrollToEnd {
		t.Fatal("a tool event forced the transcript back to the end")
	}
}

// The transcript must reserve the room the composer actually covers, so a
// tall draft or an error line never lands on the last message.
func TestComposerInsetFollowsItsHeight(t *testing.T) {
	_, a := seed(t)
	a.home = false
	a.draft = strings.Repeat("这是一行很长的草稿，用来把输入框撑到六行。\n", 12)
	a.errText = "模型服务返回 401：API Key 无效，请到设置里重新填写后再试一次。"
	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	settle(tt)
	tt.Frame()
	settle(tt)

	field, ok := tt.Find("输入区")
	if !ok {
		t.Fatal("composer missing")
	}
	errEl, ok := tt.Find("模型服务返回 401：API Key 无效，请到设置里重新填写后再试一次。")
	if !ok {
		t.Fatal("error line missing")
	}
	last, ok := tt.Find("小测")
	if !ok {
		t.Fatal("last message missing")
	}
	if errEl.Y+errEl.H > 820 {
		t.Fatalf("the error line falls outside the window: bottom=%v", errEl.Y+errEl.H)
	}
	if last.Y+last.H > field.Y {
		t.Fatalf("the composer covers the last message: composer top=%v, message bottom=%v",
			field.Y, last.Y+last.H)
	}
	if a.bottomInset() <= chatBottomInset {
		t.Fatalf("the inset did not grow with the composer: %v", a.bottomInset())
	}
}

// Conversations are grouped by the local day. Truncating the instant used
// the UTC day, which is eight hours off in Asia/Shanghai.
func TestConversationGroupsUseLocalMidnight(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skip("no tzdata")
	}
	now := time.Now().In(loc)
	// 00:30 local today is still "today", though its UTC instant is yesterday.
	early := time.Date(now.Year(), now.Month(), now.Day(), 0, 30, 0, 0, loc)
	groups := groupConversations([]store.Conversation{{ID: "a", Title: "凌晨", UpdatedAt: early}})
	if len(groups) != 1 || groups[0].label != "今天" {
		t.Fatalf("a conversation from 00:30 local landed in %v, want 今天", groups)
	}
}

// A picked answer must survive a restart. It used to live only in memory, so
// reopening the app showed every quiz as unanswered.
func TestQuizAnswerSurvivesReload(t *testing.T) {
	_, a := seed(t)
	a.home = false
	var quizMsg store.Message
	for _, msg := range a.messages {
		if msg.Kind == "quiz" {
			quizMsg = msg
		}
	}
	if quizMsg.ID == "" {
		t.Fatal("seed has no quiz")
	}
	var payload quizPayload
	if err := json.Unmarshal([]byte(quizMsg.Payload), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Questions) != 1 {
		t.Fatalf("want 1 question, got %d", len(payload.Questions))
	}
	a.saveQuizChoice(quizMsg, payload, 0, 2)

	// Drop the in-memory choice and read the transcript back from the store,
	// as a restart does.
	a.quizChoices = map[string]int{}
	if err := a.reloadForConversation(a.active); err != nil {
		t.Fatal(err)
	}
	var reloaded quizPayload
	for _, msg := range a.messages {
		if msg.Kind == "quiz" {
			if err := json.Unmarshal([]byte(msg.Payload), &reloaded); err != nil {
				t.Fatal(err)
			}
		}
	}
	if reloaded.Questions[0].Chosen != 2 {
		t.Fatalf("the answer was not stored: chosen=%d", reloaded.Questions[0].Chosen)
	}

	tt := ui.NewTester(a.view, 1240, 820)
	tt.Frame()
	settle(tt)
	if !tt.HasText("你的选择") {
		t.Fatal("the reloaded quiz does not show the answer that was picked")
	}
}

// An older quiz with no chosen field must not read as "option A was picked".
func TestLegacyQuizStaysUnanswered(t *testing.T) {
	if quizPayloadHasChoice(`{"questions":[{"id":"q","prompt":"p","options":["a","b"],"answerIndex":0}]}`) {
		t.Fatal("a payload without a chosen field reported a choice")
	}
	if !quizPayloadHasChoice(`{"questions":[{"id":"q","chosen":1}]}`) {
		t.Fatal("a payload with a chosen field was not detected")
	}
}
