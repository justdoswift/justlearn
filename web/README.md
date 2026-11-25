# JustLearn Web

The browser client. It talks to the same Go store and tutor engine the macOS
app uses, so a route, a lesson and a quiz mean exactly the same thing in both.

- **Client** — React 19, Tailwind CSS 4, shadcn/ui (Base UI, `nova`), with
  [Solid Glass](https://deeplethe.github.io/solid-glass) as the material.
- **Server** — `desktop/cmd/web`, a small HTTP + SSE layer over
  `desktop/internal/store` and `desktop/internal/agent`.

## Run it

Two processes in development, one in production.

```bash
# 1. the API and the tutor engine
cd desktop
go run ./cmd/web -db ./justlearn.db        # http://localhost:8787

# 2. the client
cd web
npm install
npm run dev                                 # http://localhost:5191
```

Vite proxies `/api` to `:8787`, so the client and the API share an origin and
the SSE stream needs no CORS workaround.

For production, build the client and let the Go server serve it:

```bash
cd web && npm run build
cd ../desktop && go run ./cmd/web -db ./justlearn.db
# open http://localhost:8787
```

`cmd/web` serves `web/dist` when it exists and falls back to `index.html` for
deep links. Without a build it serves the API alone and says so in the log.

## Configuration

The server reads `../.env`, the same file the desktop app reads:

| Variable | What it is |
|---|---|
| `PI_BASE_URL` | Any OpenAI-compatible endpoint. |
| `PI_MODEL` | The model id, when the settings row is empty. |
| `PI_API_KEY` | The key. The client never sees it; it only learns whether one is set. |

Settings written in the app (base URL, model, appearance) go to the same
SQLite rows the desktop app uses. Appearance also lives in `localStorage`, so
the page paints the right light before the first request lands.

## The API

Everything is under `/api`. Reads return camelCase DTOs from
`desktop/server/dto.go`; the store's own structs stay Go-shaped.

| Method | Path | What it does |
|---|---|---|
| `GET` | `/bootstrap` | Routes, conversations, active id, whether a key is set. |
| `GET` | `/goals` | The routes. |
| `POST` | `/goals` | Create a route. |
| `DELETE` | `/goals/{id}` | Delete a route and its conversations. |
| `GET` | `/conversations` | Every conversation. |
| `POST` | `/conversations` | Create one, optionally under a route. |
| `DELETE` | `/conversations/{id}` | Delete one. |
| `GET` | `/conversations/{id}/messages` | The transcript. |
| `GET` | `/goals/{id}/route` | Ordered nodes plus the current node. |
| `GET` | `/goals/{id}/abilities` | The route's abilities. |
| `POST` | `/conversations/{id}/send` | **One turn, streamed as SSE.** |
| `POST` | `/conversations/{id}/stop` | Cancel the running turn. |
| `POST` | `/messages/{id}/answer` | Record a quiz choice in the payload. |
| `GET` | `/models` | Ask the endpoint what models it has. |
| `GET` `PUT` | `/settings` | Base URL, model, appearance. |

### The turn stream

`POST /conversations/{id}/send` answers with `text/event-stream`. Each frame is
one JSON object:

```
data: {"kind":"delta","text":"零基础学"}
data: {"kind":"message","message":{...}}
data: {"kind":"tool"}
data: {"kind":"error","error":"..."}
data: {"kind":"done"}
```

`delta` is text as the model writes it. `message` is a stored message — the
reader's own, then the tutor's. `tool` means a learning tool ran and the route
or the ability profile may have changed, so the client refetches those two
without touching the transcript. Aborting the request cancels the turn
server-side, because the engine runs under the request context.

## Layout

```
web/
├── src/
│   ├── app/
│   │   ├── app.tsx              the shell: home or a conversation, the two margins
│   │   ├── home.tsx             the desk before anything is on it
│   │   ├── composer-bar.tsx     the one way in, on the home page and in a turn
│   │   ├── transcript.tsx       the conversation, and following it as it writes
│   │   ├── markdown.tsx         the reading surface
│   │   ├── lesson-card.tsx      a lesson the tutor set down for this node
│   │   ├── quiz-card.tsx        one question, answered by pressing the row
│   │   ├── canvas.tsx           the controlled blocks a tutor may draw
│   │   ├── route-rail.tsx       the roadmap as a scale in the left margin
│   │   ├── ability-rail.tsx     the ability profile as heat in the right margin
│   │   ├── nav-sheet.tsx        routes and past conversations, on a sheet
│   │   ├── settings-dialog.tsx  appearance, model, about
│   │   └── chrome.tsx           the ✦ mark, the small buttons, the wordmark
│   ├── lib/
│   │   ├── api.ts               the client, and the SSE reader
│   │   ├── course.tsx           the one store: route, transcript, turn, margins
│   │   └── types.ts             the shapes the server sends
│   ├── components/ui/           shadcn/ui, Solid Glass's own, and the rest
│   ├── solid-glass.css          the material (upstream, unmodified)
│   ├── brand.css                JustLearn's warm paper, ink and jade, on top
│   ├── reading.css              the type for prose, and a lesson's title
│   └── index.css                the import chain
└── vite.config.ts               the /api proxy
```

## The design

The material is Solid Glass, unmodified: its rim of light along a pane's top
edge, its hairline, its shadow, and its three thicknesses. `brand.css` only
retints it — warm paper instead of cool grey, ink instead of black, jade in
place of the neutral brand. The four material rules still hold:

- **The finish is at the edge.** Rim, hairline, shadow. Never a gradient
  across the face.
- **Depth is thickness.** Panel < card < popover. Thicker is more opaque.
- **One accent.** Jade marks only what is live: the current stage, the reader's
  own question, a lesson's kicker, a correct answer, the send mark.
- **One state.** Hovered, chosen and held-open wear the same clear glass.

What is JustLearn's own:

- **The page is the desk.** Nothing is boxed that does not have to be. The
  conversation is a reading column of 680px; only a lesson, a question and the
  places you write in are panes.
- **A question is set in the serif, an answer in the sans.** The reader's words
  are the heading of the exchange, marked by a jade rule; the tutor answers in
  prose. That is what keeps a transcript from reading as one grey wall.
- **The margins are margins, not panels.** The route is a scale of ticks, the
  ability profile a column of heat cells. Neither takes a border, and neither
  pushes the reading column.
- **Mastery is fill, not outline.** An outline around every cell makes an
  ability profile read as a row of checkboxes.
- **A quiz option is a row you press, not a form field.** Faint fill, no
  border; the border is what made the quiz look like a form.
- **Weight stops at 400.** Hierarchy comes from size, colour, measure and the
  serif, not from a heavier stroke.
