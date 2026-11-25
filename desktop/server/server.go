// Package server exposes the tutor engine over HTTP so a web client can use
// the same store, agent and tools the desktop app uses.
//
// The desktop app draws its own UI in Go; this package is the other half of
// the same product: a JSON API plus one SSE stream for a turn. It owns no
// learning logic of its own — every read and write goes through store and
// agent, so the two clients can never drift apart on what a route is.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/agent"
	"github.com/justdoswift/justlearn/desktop/internal/store"
)

// Server is the HTTP surface over one store.
type Server struct {
	store *store.Store
	// engineConfig is read per turn so a settings change applies without a
	// restart.
	engineConfig func() agent.Config
	// apiKey reports whether a key is configured, so the client can send the
	// reader to settings instead of failing a turn.
	apiKey func() string

	mu      sync.Mutex
	streams map[string]context.CancelFunc
}

// Config is what the server needs from its host.
type Config struct {
	Store        *store.Store
	EngineConfig func() agent.Config
	APIKey       func() string
}

// New builds a server over the given store.
func New(cfg Config) *Server {
	return &Server{
		store:        cfg.Store,
		engineConfig: cfg.EngineConfig,
		apiKey:       cfg.APIKey,
		streams:      map[string]context.CancelFunc{},
	}
}

// Handler returns the router. Every route is under /api so the static client
// can be served from the same origin without collisions.
func (s *Server) Handler(staticDir string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/bootstrap", s.handleBootstrap)
	mux.HandleFunc("GET /api/goals", s.handleGoals)
	mux.HandleFunc("POST /api/goals", s.handleCreateRoute)
	mux.HandleFunc("DELETE /api/goals/{id}", s.handleDeleteRoute)
	mux.HandleFunc("GET /api/conversations", s.handleConversations)
	mux.HandleFunc("POST /api/conversations", s.handleCreateConversation)
	mux.HandleFunc("DELETE /api/conversations/{id}", s.handleDeleteConversation)
	mux.HandleFunc("GET /api/conversations/{id}/messages", s.handleMessages)
	mux.HandleFunc("GET /api/goals/{id}/route", s.handleRoute)
	mux.HandleFunc("GET /api/goals/{id}/abilities", s.handleAbilities)
	mux.HandleFunc("POST /api/conversations/{id}/send", s.handleSend)
	mux.HandleFunc("POST /api/conversations/{id}/stop", s.handleStop)
	mux.HandleFunc("POST /api/messages/{id}/answer", s.handleAnswer)
	mux.HandleFunc("GET /api/models", s.handleModels)
	mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", s.handlePutSettings)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	if staticDir != "" {
		mux.Handle("/", spaHandler(staticDir))
	}
	return withCORS(mux)
}

// withCORS lets the Vite dev server on another port talk to the API while
// developing. In production both are served from one origin.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// spaHandler serves the built client and falls back to index.html so a deep
// link does not 404.
//
// index.html is never cached: its asset names change on every build, and a
// cached copy keeps pointing at the previous build's CSS and JS. The hashed
// assets under /assets are immutable and cached hard.
func spaHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-store")
			}
			files.ServeHTTP(w, r)
			return
		}
		index := filepath.Join(dir, "index.html")
		if _, err := os.Stat(index); err != nil {
			http.Error(w, "client not built: run npm run build in web/", http.StatusNotFound)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, index)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	return dec.Decode(v)
}

// ---- bootstrap ------------------------------------------------------------

// bootstrapPayload is everything the client needs for a first paint: the
// routes, the conversations, and which conversation to open.
type bootstrapPayload struct {
	Goals         []goalDTO         `json:"goals"`
	Conversations []conversationDTO `json:"conversations"`
	ActiveID      string            `json:"activeId"`
	HasKey        bool              `json:"hasKey"`
	Model         string            `json:"model"`
	Appearance    string            `json:"appearance"`
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	goals, err := s.store.Goals()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	convs, err := s.store.Conversations()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	activeID := ""
	if len(convs) > 0 {
		activeID = convs[0].ID
	}
	writeJSON(w, http.StatusOK, bootstrapPayload{
		Goals:         toGoals(goals),
		Conversations: toConversations(convs),
		ActiveID:      activeID,
		HasKey:        s.apiKey != nil && strings.TrimSpace(s.apiKey()) != "",
		Model:         s.store.GetSetting("model", ""),
		Appearance:    s.store.GetSetting("appearance", "dark"),
	})
}

// ---- routes ---------------------------------------------------------------

func (s *Server) handleGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := s.store.Goals()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, toGoals(goals))
}

func (s *Server) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	_ = readJSON(r, &body)
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = "新的学习路线"
	}
	goal, conv, err := s.store.CreateRoute(title)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"goal":         toGoal(goal),
		"conversation": toConversation(conv),
	})
}

func (s *Server) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteRoute(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---- conversations --------------------------------------------------------

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	convs, err := s.store.Conversations()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, toConversations(convs))
}

func (s *Server) handleCreateConversation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GoalID string `json:"goalId"`
	}
	_ = readJSON(r, &body)
	conv, err := s.store.CreateConversation(body.GoalID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	_ = s.store.TouchConversation(conv.ID)
	writeJSON(w, http.StatusOK, toConversation(conv))
}

func (s *Server) handleDeleteConversation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteConversation(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	msgs, err := s.store.Messages(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, toMessages(msgs))
}

// ---- route and abilities --------------------------------------------------

// routePayload is the roadmap as the client draws it: ordered nodes with the
// stage each belongs to, plus the node the reader is on.
type routePayload struct {
	Goal    goalDTO   `json:"goal"`
	Nodes   []nodeDTO `json:"nodes"`
	Current nodeDTO   `json:"current"`
}

func (s *Server) handleRoute(w http.ResponseWriter, r *http.Request) {
	goalID := r.PathValue("id")
	goal, err := s.store.Goal(goalID)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	nodes, err := s.store.Nodes(goalID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	current, _ := s.store.CurrentNode(goalID)
	writeJSON(w, http.StatusOK, routePayload{
		Goal:    toGoal(goal),
		Nodes:   toNodes(nodes),
		Current: toNode(current),
	})
}

func (s *Server) handleAbilities(w http.ResponseWriter, r *http.Request) {
	goalID := r.PathValue("id")
	abilities, err := s.store.Abilities(goalID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, toAbilities(abilities))
}

// ---- the turn -------------------------------------------------------------

// sendRequest is one reader message.
type sendRequest struct {
	Text string `json:"text"`
}

// sseEvent is one frame on the turn stream. Kind mirrors agent.Event.Kind and
// adds two the server owns: "error" and "done".
type sseEvent struct {
	Kind    string      `json:"kind"`
	Text    string      `json:"text,omitempty"`
	Message *messageDTO `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// handleSend runs one tutor turn and streams it as Server-Sent Events.
//
// The stream is the turn: every delta, message and tool event the engine
// emits is forwarded as it happens, so the client renders text as it arrives
// instead of waiting for the whole answer.
func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	convID := r.PathValue("id")
	var body sendRequest
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	text := strings.TrimSpace(body.Text)
	if text == "" {
		writeErr(w, http.StatusBadRequest, errors.New("empty message"))
		return
	}
	if s.apiKey != nil && strings.TrimSpace(s.apiKey()) == "" {
		writeErr(w, http.StatusPreconditionFailed, errors.New("no API key configured"))
		return
	}
	conv, err := s.store.Conversation(convID)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, errors.New("streaming unsupported"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	s.trackStream(convID, cancel)
	defer s.untrackStream(convID)

	send := func(ev sseEvent) bool {
		data, err := json.Marshal(ev)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	engine := &agent.Engine{Store: s.store, Config: s.engineConfig()}
	runErr := engine.Send(ctx, conv, text, func(ev agent.Event) {
		switch ev.Kind {
		case "delta":
			send(sseEvent{Kind: "delta", Text: ev.Text})
		case "message":
			if ev.Message != nil {
				m := toMessage(*ev.Message)
				send(sseEvent{Kind: "message", Message: &m})
			}
		case "tool":
			send(sseEvent{Kind: "tool"})
		}
	})
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		send(sseEvent{Kind: "error", Error: runErr.Error()})
	}
	send(sseEvent{Kind: "done"})
}

func (s *Server) trackStream(convID string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if prev, ok := s.streams[convID]; ok {
		prev()
	}
	s.streams[convID] = cancel
}

func (s *Server) untrackStream(convID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.streams, convID)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	convID := r.PathValue("id")
	s.mu.Lock()
	cancel, ok := s.streams[convID]
	s.mu.Unlock()
	if ok {
		cancel()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---- quiz answers ---------------------------------------------------------

// handleAnswer records the option a reader picked on a saved quiz. The answer
// lives in the message payload, so it survives a reload.
func (s *Server) handleAnswer(w http.ResponseWriter, r *http.Request) {
	msgID := r.PathValue("id")
	var body struct {
		Question int `json:"question"`
		Choice   int `json:"choice"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	msgs, err := s.store.MessageByID(msgID)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(msgs.Payload), &payload); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	questions, _ := payload["questions"].([]any)
	if body.Question < 0 || body.Question >= len(questions) {
		writeErr(w, http.StatusBadRequest, errors.New("question out of range"))
		return
	}
	q, _ := questions[body.Question].(map[string]any)
	if q == nil {
		writeErr(w, http.StatusBadRequest, errors.New("question malformed"))
		return
	}
	q["chosen"] = body.Choice
	data, err := json.Marshal(payload)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.store.UpdateMessagePayload(msgID, string(data)); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---- models and settings --------------------------------------------------

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	models, err := agent.ListModels(ctx, s.engineConfig())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"models": []string{}, "error": err.Error()})
		return
	}
	if models == nil {
		models = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"models": models})
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"baseUrl":    s.store.GetSetting("base_url", ""),
		"model":      s.store.GetSetting("model", ""),
		"appearance": s.store.GetSetting("appearance", "dark"),
		"hasKey":     s.apiKey != nil && strings.TrimSpace(s.apiKey()) != "",
	})
}

func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BaseURL    *string `json:"baseUrl"`
		Model      *string `json:"model"`
		Appearance *string `json:"appearance"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if body.BaseURL != nil {
		if err := s.store.SetSetting("base_url", strings.TrimSpace(*body.BaseURL)); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
	}
	if body.Model != nil {
		if err := s.store.SetSetting("model", strings.TrimSpace(*body.Model)); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
	}
	if body.Appearance != nil {
		if err := s.store.SetSetting("appearance", strings.TrimSpace(*body.Appearance)); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
