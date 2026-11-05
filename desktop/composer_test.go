package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

func cropComposer(img image.Image) image.Image {
	return img.(*image.RGBA).SubImage(image.Rect(648, 1590, 2072, 1800))
}

func cropTall(img image.Image) image.Image {
	return img.(*image.RGBA).SubImage(image.Rect(648, 1350, 2072, 1800))
}

func savePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestRenderComposer(t *testing.T) {
	dir := filepath.Join("..", ".oil-ui", "composer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, a := seed(t)
	render := func(name string, prep func()) {
		prep()
		savePNG(t, filepath.Join(dir, name+".png"), cropComposer(ui.Render(a.view, 1440, 900, 2)))
	}

	render("rest", func() { a.draft = "" })
	render("typing", func() { a.draft = "为什么主键要选学号，而不是姓名？" })
	render("multiline", func() {
		a.draft = "第一，学号是固定的；\n第二，姓名可能重复；\n第三，主键不能随业务变化。"
	})
	render("streaming", func() {
		a.draft = ""
		a.streaming = true
		a.streamText = "先别急着写答案，先说说你判断主键的两条标准。"
	})
	render("error", func() {
		a.streaming = false
		a.errText = "模型服务返回 401：API Key 无效"
	})

	// Focused state: focus the field and let the next frame read it.
	a.errText = ""
	a.draft = ""
	a.focusDraft = true
	tt := ui.NewTester(a.view, 1440, 900)
	tt.SetScale(2)
	tt.Frame()
	tt.Frame()
	savePNG(t, filepath.Join(dir, "focused.png"), cropComposer(tt.Image()))

	// Model menu, with a few choices and the current model.
	a.modelMenuOpen = true
	a.modelsLoaded = true
	a.models = []string{"deepseek-v4.1-flash"}
	savePNG(t, filepath.Join(dir, "model.png"), cropTall(ui.Render(a.view, 1440, 900, 2)))

	// Empty route state: the placeholder changes.
	st, err := store.Open(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EnsureSeed(); err != nil {
		t.Fatal(err)
	}
	_ = st.SetSetting("model", "deepseek-v4.1-flash")
	empty := newApp(st)
	savePNG(t, filepath.Join(dir, "empty.png"), cropComposer(ui.Render(empty.view, 1440, 900, 2)))
}

func TestRenderComposerDark(t *testing.T) {
	dir := filepath.Join("..", ".oil-ui", "composer", "dark")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, a := seed(t)
	tt := ui.NewTester(a.view, 1440, 900)
	tt.SetScale(2)
	tt.SetDark(true)
	render := func(name string, prep func()) {
		prep()
		tt.Frame()
		savePNG(t, filepath.Join(dir, name+".png"), cropComposer(tt.Image()))
	}
	render("rest", func() { a.draft = "" })
	render("typing", func() { a.draft = "为什么主键要选学号，而不是姓名？" })
	render("multiline", func() {
		a.draft = "第一，学号是固定的；\n第二，姓名可能重复；\n第三，主键不能随业务变化。"
	})
	render("streaming", func() {
		a.draft = ""
		a.streaming = true
	})
	render("error", func() {
		a.streaming = false
		a.errText = "模型服务返回 401：API Key 无效"
	})
	a.errText = ""
	a.draft = ""
	a.focusDraft = true
	tt.Frame()
	tt.Frame()
	savePNG(t, filepath.Join(dir, "focused.png"), cropComposer(tt.Image()))

	a.modelMenuOpen = true
	a.modelsLoaded = true
	a.models = []string{"deepseek-v4.1-flash"}
	tt.Frame()
	savePNG(t, filepath.Join(dir, "model.png"), cropTall(tt.Image()))
}

func TestRenderModelMenuFull(t *testing.T) {
	_, a := seed(t)
	a.modelMenuOpen = true
	a.modelsLoaded = true
	a.models = []string{"deepseek-v4.1-flash"}
	dir := filepath.Join("..", ".oil-ui", "composer")
	savePNG(t, filepath.Join(dir, "model-full.png"), ui.Render(a.view, 1440, 900, 2))
}
