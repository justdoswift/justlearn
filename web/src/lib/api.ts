import type {
  Ability,
  Bootstrap,
  Conversation,
  Goal,
  Message,
  Route,
  Settings,
  TurnEvent,
} from "./types"

const BASE = "/api"

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: init?.body ? { "Content-Type": "application/json", ...init?.headers } : init?.headers,
  })
  if (!res.ok) {
    let detail = res.statusText
    try {
      const body = await res.json()
      if (typeof body?.error === "string") detail = body.error
    } catch {
      // The body was not JSON; the status line is all there is.
    }
    throw new Error(detail)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export const api = {
  bootstrap: () => request<Bootstrap>("/bootstrap"),

  goals: () => request<Goal[]>("/goals"),

  createRoute: (title: string) =>
    request<{ goal: Goal; conversation: Conversation }>("/goals", {
      method: "POST",
      body: JSON.stringify({ title }),
    }),

  deleteRoute: (id: string) => request<{ ok: boolean }>(`/goals/${id}`, { method: "DELETE" }),

  conversations: () => request<Conversation[]>("/conversations"),

  createConversation: (goalId: string) =>
    request<Conversation>("/conversations", { method: "POST", body: JSON.stringify({ goalId }) }),

  deleteConversation: (id: string) =>
    request<{ ok: boolean }>(`/conversations/${id}`, { method: "DELETE" }),

  messages: (id: string) => request<Message[]>(`/conversations/${id}/messages`),

  route: (goalId: string) => request<Route>(`/goals/${goalId}/route`),

  abilities: (goalId: string) => request<Ability[]>(`/goals/${goalId}/abilities`),

  answer: (messageId: string, question: number, choice: number) =>
    request<{ ok: boolean }>(`/messages/${messageId}/answer`, {
      method: "POST",
      body: JSON.stringify({ question, choice }),
    }),

  settings: () => request<Settings>("/settings"),

  saveSettings: (patch: Partial<Pick<Settings, "baseUrl" | "model" | "appearance">>) =>
    request<{ ok: boolean }>("/settings", { method: "PUT", body: JSON.stringify(patch) }),

  models: () => request<{ models: string[]; error?: string }>("/models"),

  stop: (conversationId: string) =>
    request<{ ok: boolean }>(`/conversations/${conversationId}/stop`, { method: "POST" }),
}

/**
 * One turn, read off the server's SSE stream as it is written. The stream is
 * the turn: deltas arrive as the model writes them, and a `message` event
 * carries whatever was stored.
 */
export async function sendTurn(
  conversationId: string,
  text: string,
  onEvent: (event: TurnEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const res = await fetch(`${BASE}/conversations/${conversationId}/send`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text }),
    signal,
  })

  if (!res.ok) {
    let detail = res.statusText
    try {
      const body = await res.json()
      if (typeof body?.error === "string") detail = body.error
    } catch {
      // Not JSON; keep the status line.
    }
    throw new Error(detail)
  }
  if (!res.body) throw new Error("这个浏览器读不了流式响应。")

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ""

  const drain = (chunk: string) => {
    buffer += chunk
    let cut = buffer.indexOf("\n\n")
    while (cut >= 0) {
      const frame = buffer.slice(0, cut)
      buffer = buffer.slice(cut + 2)
      for (const line of frame.split("\n")) {
        if (!line.startsWith("data:")) continue
        const raw = line.slice(5).trim()
        if (!raw) continue
        try {
          onEvent(JSON.parse(raw) as TurnEvent)
        } catch {
          // A frame that is not JSON is not ours to render.
        }
      }
      cut = buffer.indexOf("\n\n")
    }
  }

  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    drain(decoder.decode(value, { stream: true }))
  }
  drain(decoder.decode())
}
