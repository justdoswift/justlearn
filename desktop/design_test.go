package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

func renderSettled(t *testing.T, a *app, path string, w, h int) {
	t.Helper()
	tt := ui.NewTester(a.view, w, h)
	tt.SetScale(2)
	tt.Frame()
	tt.Move(float32(w-2), float32(h-2))
	tt.Advance(700 * time.Millisecond)
	savePNG(t, path, tt.Image())
}

// TestRenderDesign renders the home and chat states at 1440x900.
// Set OIL_RENDER to the output name.
func TestRenderDesign(t *testing.T) {
	name := os.Getenv("OIL_RENDER")
	if name == "" {
		name = "current"
	}
	_, a := seed(t)
	dir := filepath.Join("..", ".oil-ui", "redesign")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	renderSettled(t, a, filepath.Join(dir, name+"-home.png"), 1440, 900)

	light := *a
	light.dark = false
	light.appearance = "light"
	renderSettled(t, &light, filepath.Join(dir, name+"-home-light.png"), 1440, 900)

	a.home = false
	renderSettled(t, a, filepath.Join(dir, name+"-chat.png"), 1440, 900)

	light.home = false
	renderSettled(t, &light, filepath.Join(dir, name+"-chat-light.png"), 1440, 900)

	a.toggleSettingsPanel()
	renderSettled(t, a, filepath.Join(dir, name+"-blank-settings.png"), 1440, 900)
	a.closePanel()
	a.panelShown = 0

	light.home = false
	light.toggleSettingsPanel()
	renderSettled(t, &light, filepath.Join(dir, name+"-blank-settings-light.png"), 1440, 900)
}

func TestRenderStates(t *testing.T) {
	dir := filepath.Join("..", ".oil-ui", "redesign")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Empty app: the home composer is the whole screen.
	st, err := store.Open(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	empty := newApp(st)
	renderSettled(t, empty, filepath.Join(dir, "empty.png"), 1440, 900)

	// Answered quiz, taller than the window.
	_, a := seed(t)
	a.home = false
	for _, msg := range a.messages {
		if msg.Kind != "quiz" {
			continue
		}
		var qp quizPayload
		_ = json.Unmarshal([]byte(msg.Payload), &qp)
		if len(qp.Questions) > 0 {
			a.quizChoices[qp.Questions[0].ID] = qp.Questions[0].AnswerIndex
		}
	}
	a.chatList.FollowEnd = false
	a.chatList.FollowAnimation = nil
	renderSettled(t, a, filepath.Join(dir, "quiz.png"), 1440, 1280)

	// Functional settings fallback, reached from the menu or first run.
	_, stg := seed(t)
	stg.openSettingsAt(1)
	renderSettled(t, stg, filepath.Join(dir, "settings-functional.png"), 1440, 900)

	// Streaming.
	_, s := seed(t)
	s.home = false
	s.streaming = true
	s.streamText = "先看表结构：一行是一条记录，一列是一个字段。"
	renderSettled(t, s, filepath.Join(dir, "streaming.png"), 1440, 900)

	// Narrow home and chat.
	_, n := seed(t)
	renderSettled(t, n, filepath.Join(dir, "narrow-home.png"), 1120, 740)
	n.home = false
	n.chatList.FollowEnd = false
	n.chatList.FollowAnimation = nil
	renderSettled(t, n, filepath.Join(dir, "narrow-chat.png"), 1120, 740)
}
