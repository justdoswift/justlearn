package main

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestRenderDesign renders the workspace at the standard 1440x900 viewport.
// Set OIL_RENDER to the output name, e.g. before or after.
func TestRenderDesign(t *testing.T) {
	name := os.Getenv("OIL_RENDER")
	if name == "" {
		name = "current"
	}
	_, a := seed(t)
	if os.Getenv("OIL_RENDER_TOP") != "" {
		a.chatList.FollowEnd = false
		a.chatList.FollowAnimation = nil
	}
	dir := filepath.Join("..", ".oil-ui", "redesign")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	img := ui.Render(a.view, 1440, 900, 2)
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestRenderStates(t *testing.T) {
	_, a := seed(t)
	dir := filepath.Join("..", ".oil-ui", "redesign")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Settings dialog.
	a.settingsOpen = true
	img := ui.Render(a.view, 1440, 900, 2)
	f, err := os.Create(filepath.Join(dir, "after-settings.png"))
	if err != nil {
		t.Fatal(err)
	}
	_ = png.Encode(f, img)
	f.Close()
	a.settingsOpen = false

	// Dark appearance.
	tt := ui.NewTester(a.view, 1440, 900)
	tt.SetDark(true)
	tt.Frame()
	f, err = os.Create(filepath.Join(dir, "after-dark.png"))
	if err != nil {
		t.Fatal(err)
	}
	_ = png.Encode(f, tt.Image())
	f.Close()

	// Quiz answered state, taller than the window so both cards show.
	for i, msg := range a.messages {
		if msg.Kind != "quiz" {
			continue
		}
		var qp quizPayload
		_ = json.Unmarshal([]byte(msg.Payload), &qp)
		if len(qp.Questions) > 0 {
			a.quizChoices[qp.Questions[0].ID] = qp.Questions[0].AnswerIndex
		}
		_ = i
	}
	a.chatList.FollowEnd = false
	a.chatList.FollowAnimation = nil
	img = ui.Render(a.view, 1440, 1280, 2)
	f, err = os.Create(filepath.Join(dir, "after-quiz.png"))
	if err != nil {
		t.Fatal(err)
	}
	_ = png.Encode(f, img)
	f.Close()
}

func TestRenderNarrow(t *testing.T) {
	_, a := seed(t)
	dir := filepath.Join("..", ".oil-ui", "redesign")
	img := ui.Render(a.view, 1120, 740, 2)
	f, err := os.Create(filepath.Join(dir, "after-narrow.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestRenderStreaming(t *testing.T) {
	_, a := seed(t)
	a.streaming = true
	a.streamText = "先看表结构：一行是一条记录，一列是一个字段。"
	dir := filepath.Join("..", ".oil-ui", "redesign")
	img := ui.Render(a.view, 1440, 900, 2)
	f, err := os.Create(filepath.Join(dir, "after-streaming.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
