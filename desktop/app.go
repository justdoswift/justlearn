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

	goals     []store.Goal
	convs     []store.Conversation
	active    store.Conversation
	goal      store.Goal
	nodes     []store.Node
	abilities []store.Ability
	current   store.Node
	currents  map[string]string

	messages   []store.Message
	streaming  bool
	streamText string
	errText    string

	draft      string
	md         *markdown.Options
	chatList   ui.ListState
	composer   ui.Handle
	focusDraft bool
	// chatScrollToEnd asks the next chat frame to scroll from the view,
	// where ListState.ScrollToEnd applies.
	chatScrollToEnd bool

	home          bool
	railSelected  int
	railShown     float32
	heatShown     float32
	heatHover     int
	heatCardShown float32

	panelOpen  bool
	panelShown float32

	settingsOpen     bool
	settingsShown    float32
	settingsSection  int
	settingsSaved    bool
	appearance       string
	dark             bool
	systemDark       bool
	version          string
	search           string
	confirmRouteOpen bool
	pendingRouteID   string
	confirmChatOpen  bool
	pendingChatID    string
	modelMenuOpen    bool
	models           []string
	modelsLoaded     bool
	modelsLoading    bool
	modelsErr        string
	formBase         string
	formModel        string
	formKey          string

	quizChoices  map[string]int
	flippedCards map[string]bool
	expandedCode map[string]bool
	sendCancel   context.CancelFunc
}

func newApp(st *store.Store) *app {
	a := &app{
		store:        st,
		quizChoices:  map[string]int{},
		flippedCards: map[string]bool{},
		expandedCode: map[string]bool{},
		appearance:   "system",
		dark:         true,
		systemDark:   true,
		home:         true,
		railSelected: -1,
		heatHover:    -1,
	}
	a.md = &markdown.Options{ShareSelection: true}
	a.chatList.FollowEnd = true
	a.chatList.FollowAnimation = &ui.Smooth
	a.chatList.Key = func(row int) any {
		if row == 0 {
			return "spacer-top"
		}
		msg := row - 1
		if msg < len(a.messages) {
			return a.messages[msg].ID
		}
		if msg == len(a.messages) && a.streaming {
			return "live"
		}
		return "spacer-bottom"
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

func (a *app) askDeleteRoute(id string) {
	if a.streaming {
		a.errText = "先停止当前回答，再删除路线。"
		return
	}
	a.pendingRouteID = id
	a.confirmRouteOpen = true
	a.confirmChatOpen = false
	a.panelOpen = false
}

func (a *app) askDeleteConversation(id string) {
	if a.streaming {
		a.errText = "先停止当前回答，再删除对话。"
		return
	}
	a.pendingChatID = id
	a.confirmChatOpen = true
	a.confirmRouteOpen = false
}

func (a *app) deleteRoute(id string) {
	if id == "" {
		return
	}
	if err := a.store.DeleteRoute(id); err != nil {
		a.errText = err.Error()
		a.confirmRouteOpen = false
		return
	}
	a.confirmRouteOpen = false
	a.pendingRouteID = ""
	activeDeleted := a.active.GoalID == id || a.goal.ID == id
	goals, _ := a.store.Goals()
	a.goals = goals
	if len(goals) == 0 {
		a.active = store.Conversation{}
		a.goal = store.Goal{}
		a.nodes = nil
		a.abilities = nil
		a.current = store.Node{}
		a.messages = nil
		a.convs = nil
		a.newRoute()
		return
	}
	if activeDeleted {
		a.active = store.Conversation{}
		a.goal = store.Goal{}
		a.nodes = nil
		a.abilities = nil
		a.current = store.Node{}
		a.messages = nil
		convs, _ := a.store.Conversations()
		var chosen store.Conversation
		for _, c := range convs {
			if c.GoalID == goals[0].ID {
				chosen = c
				break
			}
		}
		if chosen.ID == "" {
			chosen, _ = a.store.CreateConversation(goals[0].ID)
		}
		a.active = chosen
	}
	_ = a.reloadAll()
	if a.active.ID == "" {
		a.selectGoal(a.goals[0].ID)
	}
}

func (a *app) deleteConversation(id string) {
	if id == "" {
		return
	}
	if err := a.store.DeleteConversation(id); err != nil {
		a.errText = err.Error()
		a.confirmChatOpen = false
		return
	}
	a.confirmChatOpen = false
	a.pendingChatID = ""
	if a.active.ID == id {
		a.active = store.Conversation{}
		convs, _ := a.store.Conversations()
		var chosen store.Conversation
		for _, c := range convs {
			if c.GoalID == a.goal.ID {
				chosen = c
				break
			}
		}
		if chosen.ID == "" {
			chosen, _ = a.store.CreateConversation(a.goal.ID)
		}
		a.active = chosen
	}
	_ = a.reloadAll()
	if a.active.ID == "" {
		if len(a.goals) > 0 {
			a.selectGoal(a.goals[0].ID)
		} else {
			a.newRoute()
		}
	}
}

func (a *app) closeSettings() {
	a.settingsOpen = false
	a.settingsSaved = false
	a.focusDraft = true
}

func (a *app) openSettings() {
	a.openSettingsAt(0)
}

func (a *app) openSettingsAt(section int) {
	a.panelOpen = false
	a.settingsOpen = true
	a.settingsSaved = false
	if section < 0 || section > 2 {
		section = 0
	}
	a.settingsSection = section
	a.formBase = cfgBaseURL(a.store)
	a.formModel = cfgModel(a.store)
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

func (a *app) visibleNodes() []store.Node {
	q := strings.ToLower(strings.TrimSpace(a.search))
	if q == "" {
		return a.nodes
	}
	out := make([]store.Node, 0, len(a.nodes))
	for _, node := range a.nodes {
		if strings.Contains(strings.ToLower(node.Title), q) {
			out = append(out, node)
		}
	}
	return out
}

func (a *app) visibleRoutes() []store.Goal {
	q := strings.ToLower(strings.TrimSpace(a.search))
	if q == "" {
		return a.goals
	}
	out := make([]store.Goal, 0, len(a.goals))
	for _, goal := range a.goals {
		if strings.Contains(strings.ToLower(goal.Title), q) {
			out = append(out, goal)
		}
	}
	return out
}

func (a *app) visibleConversations() []store.Conversation {
	q := strings.ToLower(strings.TrimSpace(a.search))
	if q == "" {
		return a.convs
	}
	out := make([]store.Conversation, 0, len(a.convs))
	for _, conv := range a.convs {
		if strings.Contains(strings.ToLower(conv.Title), q) {
			out = append(out, conv)
		}
	}
	return out
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
	a.nodes = nil
	a.abilities = nil
	a.current = store.Node{}
	a.railSelected = -1
	a.heatHover = -1
	a.chatScrollToEnd = true
	if conv.GoalID != "" {
		if goal, err := a.store.Goal(conv.GoalID); err == nil {
			a.goal = goal
			nodes, _ := a.store.Nodes(goal.ID)
			a.nodes = nodes
			abilities, _ := a.store.Abilities(goal.ID)
			a.abilities = abilities
			if node, err := a.store.CurrentNode(goal.ID); err == nil {
				a.current = node
			}
		}
	}
	return nil
}

func (a *app) selectConversation(id string) {
	for _, conv := range a.convs {
		if conv.ID == id {
			a.active = conv
			_ = a.reloadForConversation(conv)
			a.home = len(a.messages) == 0 && !a.streaming
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
	_ = a.reloadAll()
	a.selectConversation(conv.ID)
	a.home = true
	a.focusDraft = true
}

func (a *app) newRoute() {
	a.home = true
	a.panelOpen = false
	a.modelMenuOpen = false
	goal, conv, err := a.store.CreateRoute("新的学习路线")
	if err != nil {
		a.errText = err.Error()
		return
	}
	a.goal = goal
	a.active = conv
	a.nodes = nil
	a.abilities = nil
	a.current = store.Node{}
	a.messages = nil
	_ = a.reloadAll()
	a.selectConversation(conv.ID)
	a.home = true
	a.focusDraft = true
}

func (a *app) selectGoal(goalID string) {
	for _, conv := range a.convs {
		if conv.GoalID == goalID {
			a.selectConversation(conv.ID)
			return
		}
	}
	conv, err := a.store.CreateConversation(goalID)
	if err != nil {
		a.errText = err.Error()
		return
	}
	a.active = conv
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
		a.openSettingsAt(1)
		a.errText = "先在设置里填入模型 API Key。"
		return
	}
	a.home = false
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

func (a *app) saveSettings() {
	base := strings.TrimSpace(a.formBase)
	model := strings.TrimSpace(a.formModel)
	_ = a.store.SetSetting("base_url", base)
	_ = a.store.SetSetting("model", model)
	key := strings.TrimSpace(a.formKey)
	if key != "" {
		_ = mygo.Secrets.Set("api-key", []byte(key))
	}
	a.settingsSaved = true
	a.formKey = ""
	a.errText = ""
}

func (a *app) goHome() {
	a.home = true
	a.panelOpen = false
	a.settingsOpen = false
	a.modelMenuOpen = false
	a.focusDraft = true
}

func (a *app) continueChat() {
	if len(a.messages) == 0 && !a.streaming {
		return
	}
	a.home = false
	a.focusDraft = true
	a.chatScrollToEnd = true
}

// applyAppearance saves and follows the chosen appearance. It is safe in
// headless tests: only a live window talks to the system theme.
func (a *app) applyAppearance(value string) {
	switch value {
	case "light", "dark":
	default:
		value = "system"
	}
	a.appearance = value
	_ = a.store.SetSetting("appearance", value)
	if a.win != nil {
		mygo.Theme.SetSource(mygo.ThemeSource(value))
		a.systemDark = mygo.Theme.IsDark()
	}
	switch value {
	case "light":
		a.dark = false
	case "dark":
		a.dark = true
	default:
		a.dark = a.systemDark
	}
}

// syncSystemAppearance updates the effective palette after the desktop
// switches between light and dark.
func (a *app) syncSystemAppearance() {
	if a.win != nil {
		a.systemDark = mygo.Theme.IsDark()
	}
	if a.appearance == "" || a.appearance == "system" {
		a.dark = a.systemDark
	}
}

// build 2026-11 — native client.

func (a *app) view(c *ui.Context) {
	t := appTheme(a.dark)
	p := paletteFor(a.dark)
	ui.Themed(c, t).Fill().Children(func() {
		ui.Box(c).Fill().Gradient(p.backdropA, p.backdropB, 135).Children(func() {
			a.main(c, t)
		})
		if !a.home && !a.panelOpen && !a.settingsOpen {
			a.routeRail(c, t)
		}
		a.heatPanel(c, t)
		a.overlayPanel(c, t)
		// The toolbar sits above the panels so its controls stay reachable.
		a.toolbar(c, t)
		a.settingsPage(c, t)
		a.deleteDialogs(c, t)
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
