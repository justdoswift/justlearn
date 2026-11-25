// Command web serves JustLearn to a browser: the same store, agent and tools
// the desktop app uses, over HTTP, with the built client served from the same
// origin.
//
//	go run ./cmd/web            # http://localhost:8787
//	go run ./cmd/web -addr :9000
package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/justdoswift/justlearn/desktop/internal/agent"
	"github.com/justdoswift/justlearn/desktop/internal/store"
	"github.com/justdoswift/justlearn/desktop/server"
)

func main() {
	addr := flag.String("addr", ":8787", "address to listen on")
	dbPath := flag.String("db", "", "path to the SQLite file (default: ./justlearn.db)")
	webDir := flag.String("web", "../web/dist", "directory holding the built client")
	flag.Parse()

	loadDotEnv()

	path := *dbPath
	if path == "" {
		path = "justlearn.db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		log.Fatalf("data dir: %v", err)
	}
	st, err := store.Open(path)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()
	if err := st.EnsureSeed(); err != nil {
		log.Fatalf("seed: %v", err)
	}

	srv := server.New(server.Config{
		Store:        st,
		EngineConfig: func() agent.Config { return engineConfig(st) },
		APIKey:       func() string { return apiKey() },
	})

	dir := *webDir
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		log.Printf("client not built at %s — serving the API only", dir)
		dir = ""
	} else {
		log.Printf("serving client from %s", dir)
	}

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           srv.Handler(dir),
		ReadHeaderTimeout: 10 * time.Second,
		// A turn streams for as long as the model takes, so the write
		// deadline has to stay open.
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}
	log.Printf("JustLearn web on http://localhost%s", *addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv reads ../.env, the same file the desktop app reads, so one
// configuration drives both clients.
func loadDotEnv() {
	f, err := os.Open(filepath.Join("..", ".env"))
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

func engineConfig(st *store.Store) agent.Config {
	return agent.Config{
		BaseURL:       envOr("PI_BASE_URL", st.GetSetting("base_url", "")),
		Model:         modelFor(st),
		APIKey:        apiKey(),
		UserAgent:     os.Getenv("PI_USER_AGENT"),
		SessionHeader: os.Getenv("PI_SESSION_HEADER"),
	}
}

func modelFor(st *store.Store) string {
	if model := st.GetSetting("model", ""); model != "" {
		return model
	}
	return os.Getenv("PI_MODEL")
}

func apiKey() string {
	return os.Getenv("PI_API_KEY")
}
