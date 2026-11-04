package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/agent"
	"github.com/justdoswift/justlearn/desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

type app struct {
	store  *store.Store
	engine *agent.Engine
	win    *mygo.Window

	goals    []store.Goal
	convs    []store.Conversation
	active   store.Conversation
	goal     store.Goal
	nodes    []store.Node
	current  store.Node
	currents map[string]string

	messages   []store.Message
	streaming  bool
	streamText string
	errText    string

	draft      string
	md         *markdown.Options
	chatList   ui.ListState
	composer   ui.Handle
	focusDraft bool

	routeOpen   bool
	routeNodeID string
	routeScroll ui.ScrollState

	settingsOpen  bool
	modelMenuOpen bool
	models        []string
	modelsLoaded  bool
	modelsLoading bool
	modelsErr     string
	formBase      string
	formModel     string
	formKey       string

	quizChoices map[string]int
	sendCancel  context.CancelFunc
}

func newApp(st *store.Store) *app {
	a := &app{store: st, quizChoices: map[string]int{}, routeOpen: true}
	a.md = &markdown.Options{ShareSelection: true}
	a.chatList.FollowEnd = true
	a.chatList.FollowAnimation = &ui.Smooth
	a.chatList.Key = func(row int) any {
		if row < len(a.messages) {
			return a.messages[row].ID
		}
		return "live"
	}
	a.formBase = cfgBaseURL(st)
	a.formModel = cfgModel(st)
	if err := a.reloadAll(); err != nil {
		a.errText = err.Error()
	}
	return a
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv() {
	path := filepath.Join("..", ".env")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func cfgBaseURL(st *store.Store) string {
	return envOr("PI_BASE_URL", st.GetSetting("base_url", ""))
}

func cfgModel(st *store.Store) string {
	if model := st.GetSetting("model", ""); model != "" {
		return model
	}
	return os.Getenv("PI_MODEL")
}

func prettyModel(id string) string {
	switch id {
	case "":
		return "选择模型"
	case "deepseek-v4.1-flash":
		return "DeepSeek V4.1 Flash"
	case "deepseek-v4.1":
		return "DeepSeek V4.1"
	}
	return id
}

func (a *app) loadModels() {
	if a.modelsLoading || a.modelsLoaded || a.win == nil {
		return
	}
	a.modelsLoading = true
	a.modelsErr = ""
	cfg := a.engineConfig()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		models, err := agent.ListModels(ctx, cfg)
		a.win.Update(func() {
			a.modelsLoading = false
			a.modelsLoaded = true
			if err != nil {
				a.modelsErr = err.Error()
				a.models = []string{cfg.Model}
				return
			}
			a.models = models
		})
	}()
}

func (a *app) selectModel(id string) {
	if id == "" {
		return
	}
	_ = a.store.SetSetting("model", id)
	a.formModel = id
	a.modelMenuOpen = false
	a.modelsErr = ""
}

func (a *app) attachText() {
	if a.win == nil {
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:  a.win,
			Title:   "导入文本",
			Filters: []mygo.FileFilter{{Name: "文本", Extensions: []string{"txt", "md", "markdown"}}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		data, err := os.ReadFile(paths[0])
		a.win.Update(func() {
			if err != nil {
				a.errText = "读取文件失败：" + err.Error()
				return
			}
			if len(data) > 200_000 {
				a.errText = "文件太大，请选 200KB 以内的文本"
				return
			}
			text := strings.TrimSpace(string(data))
			if text == "" {
				return
			}
			if strings.TrimSpace(a.draft) != "" {
				a.draft += "\n\n"
			}
			a.draft += text
			a.errText = ""
			a.focusDraft = true
		})
	}()
}

func (a *app) openSettings() {
	a.formBase = cfgBaseURL(a.store)
	a.formModel = cfgModel(a.store)
	a.settingsOpen = true
}

func cfgKey(st *store.Store) string {
	if key := os.Getenv("PI_API_KEY"); key != "" {
		return key
	}
	if data, err := mygo.Secrets.Get("api-key"); err == nil && len(data) > 0 {
		return string(data)
	}
	return ""
}

func (a *app) engineConfig() agent.Config {
	return agent.Config{
		BaseURL:       cfgBaseURL(a.store),
		Model:         cfgModel(a.store),
		APIKey:        cfgKey(a.store),
		UserAgent:     os.Getenv("PI_USER_AGENT"),
		SessionHeader: os.Getenv("PI_SESSION_HEADER"),
	}
}

func (a *app) reloadAll() error {
	goals, err := a.store.Goals()
	if err != nil {
		return err
	}
	convs, err := a.store.Conversations()
	if err != nil {
		return err
	}
	a.goals, a.convs = goals, convs
	a.currents = map[string]string{}
	for _, g := range goals {
		a.currents[g.ID] = g.CurrentTitle
	}
	if a.active.ID == "" {
		if len(convs) > 0 {
			a.active = convs[0]
		} else {
			conv, err := a.store.CreateConversation("")
			if err != nil {
				return err
			}
			a.active = conv
			a.convs = append([]store.Conversation{conv}, convs...)
		}
	} else {
		for _, conv := range convs {
			if conv.ID == a.active.ID {
				a.active = conv
			}
		}
	}
	return a.reloadForConversation(a.active)
}

func (a *app) reloadForConversation(conv store.Conversation) error {
	messages, err := a.store.Messages(conv.ID)
	if err != nil {
		return err
	}
	a.messages = messages
	if conv.GoalID != "" {
		if goal, err := a.store.Goal(conv.GoalID); err == nil {
			a.goal = goal
			nodes, _ := a.store.Nodes(goal.ID)
			a.nodes = nodes
			if node, err := a.store.CurrentNode(goal.ID); err == nil {
				a.current = node
			}
		}
	}
	if a.routeNodeID == "" {
		a.routeNodeID = a.current.ID
	}
	return nil
}

func (a *app) selectConversation(id string) {
	for _, conv := range a.convs {
		if conv.ID == id {
			a.active = conv
			a.routeNodeID = ""
			_ = a.reloadForConversation(conv)
			return
		}
	}
}

func (a *app) newConversation() {
	goalID := ""
	if a.goal.ID != "" {
		goalID = a.goal.ID
	}
	conv, err := a.store.CreateConversation(goalID)
	if err != nil {
		a.errText = err.Error()
		return
	}
	_ = a.store.TouchConversation(conv.ID)
	a.active = conv
	a.routeNodeID = a.current.ID
	_ = a.reloadAll()
	a.selectConversation(conv.ID)
	a.focusDraft = true
}

func (a *app) newRoute() {
	goal, conv, err := a.store.CreateRoute("新的学习路线")
	if err != nil {
		a.errText = err.Error()
		return
	}
	a.goal = goal
	a.active = conv
	a.routeNodeID = ""
	a.nodes = nil
	a.current = store.Node{}
	a.messages = nil
	_ = a.reloadAll()
	a.selectConversation(conv.ID)
	a.routeOpen = true
	a.focusDraft = true
}

func (a *app) selectGoal(goalID string) {
	for _, conv := range a.convs {
		if conv.GoalID == goalID {
			a.selectConversation(conv.ID)
			a.routeNodeID = ""
			return
		}
	}
	conv, err := a.store.CreateConversation(goalID)
	if err != nil {
		a.errText = err.Error()
		return
	}
	a.active = conv
	a.routeNodeID = ""
	_ = a.reloadAll()
	a.selectConversation(conv.ID)
}

func (a *app) send(text string) {
	text = strings.TrimSpace(text)
	if text == "" || a.streaming {
		return
	}
	if a.active.ID == "" {
		return
	}
	if cfgKey(a.store) == "" {
		a.openSettings()
		a.errText = "先在设置里填入模型 API Key。"
		return
	}
	a.draft = ""
	a.streaming = true
	a.streamText = ""
	a.errText = ""
	conv := a.active
	cfg := a.engineConfig()
	a.engine = &agent.Engine{Store: a.store, Config: cfg}
	ctx, cancel := context.WithCancel(context.Background())
	a.sendCancel = cancel
	go func() {
		err := a.engine.Send(ctx, conv, text, func(ev agent.Event) {
			a.win.Update(func() { a.handleEvent(ev) })
		})
		a.win.Update(func() { a.finishSend(err) })
	}()
}

func (a *app) handleEvent(ev agent.Event) {
	switch ev.Kind {
	case "delta":
		a.streamText += ev.Text
		a.focusDraft = false
	case "message":
		if ev.Message != nil {
			a.messages = append(a.messages, *ev.Message)
		}
	case "tool":
		_ = a.reloadAll()
	}
}

func (a *app) finishSend(err error) {
	a.streaming = false
	a.streamText = ""
	a.sendCancel = nil
	if err != nil && !errors.Is(err, context.Canceled) {
		a.errText = err.Error()
	}
	_ = a.reloadAll()
	a.focusDraft = true
}

func (a *app) cancelSend() {
	if a.sendCancel != nil {
		a.sendCancel()
		a.sendCancel = nil
	}
}

func (a *app) selectedNode() store.Node {
	id := a.routeNodeID
	if id == "" {
		id = a.current.ID
	}
	for _, node := range a.nodes {
		if node.ID == id {
			return node
		}
	}
	return a.current
}

func (a *app) saveSettings() {
	base := strings.TrimSpace(a.formBase)
	model := strings.TrimSpace(a.formModel)
	_ = a.store.SetSetting("base_url", base)
	_ = a.store.SetSetting("model", model)
	key := strings.TrimSpace(a.formKey)
	if key != "" {
		_ = mygo.Secrets.Set("api-key", []byte(key))
	}
	a.settingsOpen = false
	a.formKey = ""
	a.errText = ""
}

// build 2026-11 — native client.

func (a *app) view(c *ui.Context) {
	t := appTheme(c.Theme().Dark)
	ui.Themed(c, t).Fill().Children(func() {
		ui.Row(c).Fill().Background(t.Background).Children(func() {
			a.sidebar(c, t)
			a.main(c, t)
			if a.routeOpen {
				a.routePanel(c, t)
			}
		})
		a.settingsDialog(c, t)
	})
}

type convGroup struct {
	label string
	items []store.Conversation
}

func groupConversations(convs []store.Conversation) []convGroup {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	week := today.AddDate(0, 0, -7)
	order := []string{"今天", "昨天", "本周", "更早"}
	buckets := map[string][]store.Conversation{}
	for _, conv := range convs {
		day := conv.UpdatedAt.Local().Truncate(24 * time.Hour)
		switch {
		case day.Equal(today):
			buckets["今天"] = append(buckets["今天"], conv)
		case day.Equal(today.AddDate(0, 0, -1)):
			buckets["昨天"] = append(buckets["昨天"], conv)
		case day.After(week):
			buckets["本周"] = append(buckets["本周"], conv)
		default:
			buckets["更早"] = append(buckets["更早"], conv)
		}
	}
	var out []convGroup
	for _, label := range order {
		items := buckets[label]
		sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
		if len(items) > 0 {
			out = append(out, convGroup{label: label, items: items})
		}
	}
	return out
}

var _ = fmt.Sprint
