import * as React from "react"

import { api, sendTurn } from "./api"
import type { Ability, Appearance, Conversation, Goal, Message, Route } from "./types"

type Turn = "idle" | "thinking" | "writing" | "tool"

type CourseValue = {
  ready: boolean
  bootError: string | null

  goals: Goal[]
  conversations: Conversation[]
  activeId: string | null
  activeGoalId: string | null

  messages: Message[]
  route: Route | null
  abilities: Ability[]

  draft: string
  turn: Turn
  turnError: string | null

  hasKey: boolean
  appearance: Appearance
  model: string

  view: "home" | "chat"
  busy: boolean

  open: (id: string) => Promise<void>
  startNew: (goalId?: string) => Promise<void>
  addRoute: (title: string) => Promise<void>
  submit: (text: string) => Promise<void>
  stop: () => void
  answer: (messageId: string, question: number, choice: number) => Promise<void>
  removeConversation: (id: string) => Promise<void>
  removeRoute: (id: string) => Promise<void>
  setAppearance: (appearance: Appearance) => Promise<void>
  refresh: () => Promise<void>
}

const CourseContext = React.createContext<CourseValue | null>(null)

export function useCourse(): CourseValue {
  const value = React.useContext(CourseContext)
  if (!value) throw new Error("useCourse must be used inside <CourseProvider>")
  return value
}

function applyAppearance(appearance: Appearance) {
  const dark =
    appearance === "dark" ||
    (appearance === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches)
  document.documentElement.classList.toggle("dark", dark)
  try {
    localStorage.setItem("justlearn-appearance", appearance)
  } catch {
    // Private mode: the class is enough for this visit.
  }
}

export function CourseProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = React.useState(false)
  const [bootError, setBootError] = React.useState<string | null>(null)

  const [goals, setGoals] = React.useState<Goal[]>([])
  const [conversations, setConversations] = React.useState<Conversation[]>([])
  const [activeId, setActiveId] = React.useState<string | null>(null)
  const [activeGoalId, setActiveGoalId] = React.useState<string | null>(null)

  const [messages, setMessages] = React.useState<Message[]>([])
  const [route, setRoute] = React.useState<Route | null>(null)
  const [abilities, setAbilities] = React.useState<Ability[]>([])

  const [draft, setDraft] = React.useState("")
  const [turn, setTurn] = React.useState<Turn>("idle")
  const [turnError, setTurnError] = React.useState<string | null>(null)

  const [hasKey, setHasKey] = React.useState(true)
  const [appearance, setAppearanceState] = React.useState<Appearance>("system")
  const [model, setModel] = React.useState("")

  const abort = React.useRef<AbortController | null>(null)
  const refreshTimer = React.useRef<number | null>(null)
  const draftRef = React.useRef("")

  const clearDraft = React.useCallback(() => {
    draftRef.current = ""
    setDraft("")
  }, [])

  // ---- loading ------------------------------------------------------------

  const loadRoute = React.useCallback(async (goalId: string | null) => {
    if (!goalId) {
      setRoute(null)
      setAbilities([])
      return
    }
    try {
      const [nextRoute, nextAbilities] = await Promise.all([
        api.route(goalId),
        api.abilities(goalId),
      ])
      setRoute(nextRoute)
      setAbilities(nextAbilities)
    } catch {
      // A route that has not been drawn yet is an empty margin, not an error.
      setRoute(null)
      setAbilities([])
    }
  }, [])

  const open = React.useCallback(
    async (id: string) => {
      abort.current?.abort()
      abort.current = null
      clearDraft()
      setTurn("idle")
      setTurnError(null)

      const conversation = conversations.find((item) => item.id === id)
      setActiveId(id)
      setActiveGoalId(conversation?.goalId || null)
      try {
        const transcript = await api.messages(id)
        setMessages(transcript)
      } catch {
        setMessages([])
      }
      await loadRoute(conversation?.goalId || null)
    },
    [clearDraft, conversations, loadRoute],
  )

  const refresh = React.useCallback(async () => {
    const [nextGoals, nextConversations] = await Promise.all([api.goals(), api.conversations()])
    setGoals(nextGoals)
    setConversations(nextConversations)
    return
  }, [])

  React.useEffect(() => {
    let alive = true
    ;(async () => {
      try {
        const boot = await api.bootstrap()
        if (!alive) return
        setGoals(boot.goals)
        setConversations(boot.conversations)
        setHasKey(boot.hasKey)
        setModel(boot.model)
        setAppearanceState(boot.appearance ?? "system")
        applyAppearance(boot.appearance ?? "system")
        setReady(true)
        if (boot.activeId) {
          const conversation = boot.conversations.find((item) => item.id === boot.activeId)
          setActiveId(boot.activeId)
          setActiveGoalId(conversation?.goalId || null)
          api.messages(boot.activeId).then(setMessages).catch(() => setMessages([]))
          loadRoute(conversation?.goalId || null)
        }
      } catch (error) {
        if (!alive) return
        setBootError(error instanceof Error ? error.message : String(error))
        setReady(true)
      }
    })()
    return () => {
      alive = false
    }
  }, [loadRoute])

  // The system's own light and dark, for as long as the reader follows it.
  React.useEffect(() => {
    if (appearance !== "system") return
    const query = window.matchMedia("(prefers-color-scheme: dark)")
    const handle = () => applyAppearance("system")
    query.addEventListener("change", handle)
    return () => query.removeEventListener("change", handle)
  }, [appearance])

  // ---- starting a conversation -------------------------------------------

  const startNew = React.useCallback(
    async (goalId?: string) => {
      const target = goalId ?? activeGoalId
      if (!target) {
        setActiveId(null)
        setActiveGoalId(null)
        setMessages([])
        setRoute(null)
        setAbilities([])
        clearDraft()
        return
      }
      const conversation = await api.createConversation(target)
      setConversations((prev) => [conversation, ...prev])
      setActiveId(conversation.id)
      setActiveGoalId(target)
      setMessages([])
      clearDraft()
      setTurnError(null)
      await loadRoute(target)
    },
    [activeGoalId, clearDraft, loadRoute],
  )

  const addRoute = React.useCallback(
    async (title: string) => {
      const created = await api.createRoute(title.trim() || "新的学习路线")
      setGoals((prev) => [created.goal, ...prev])
      setConversations((prev) => [created.conversation, ...prev])
      setActiveId(created.conversation.id)
      setActiveGoalId(created.goal.id)
      setMessages([])
      clearDraft()
      await loadRoute(created.goal.id)
    },
    [clearDraft, loadRoute],
  )

  // ---- a turn -------------------------------------------------------------

  const scheduleRefresh = React.useCallback(() => {
    if (refreshTimer.current) window.clearTimeout(refreshTimer.current)
    refreshTimer.current = window.setTimeout(() => {
      if (activeGoalId) loadRoute(activeGoalId)
      refresh().catch(() => {})
    }, 240)
  }, [activeGoalId, loadRoute, refresh])

  const submit = React.useCallback(
    async (raw: string) => {
      const text = raw.trim()
      if (!text || turn !== "idle") return
      if (!hasKey) {
        setTurnError("还没有配置模型密钥。在 .env 里填上 PI_API_KEY 后重启服务。")
        return
      }

      setTurnError(null)
      clearDraft()
      setTurn("thinking")

      let conversationId = activeId

      try {
        if (!conversationId) {
          const created = await api.createRoute(text.slice(0, 24))
          conversationId = created.conversation.id
          setGoals((prev) => [created.goal, ...prev])
          setConversations((prev) => [created.conversation, ...prev])
          setActiveId(created.conversation.id)
          setActiveGoalId(created.goal.id)
          setMessages([])
        }
      } catch (error) {
        setTurnError(error instanceof Error ? error.message : String(error))
        setTurn("idle")
        return
      }

      // The reader's own words are on the page before the server answers.
      const local: Message = {
        id: `local-${Date.now()}`,
        conversationId,
        role: "user",
        kind: "text",
        text,
        payload: "",
        createdAt: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, local])

      const controller = new AbortController()
      abort.current = controller

      try {
        await sendTurn(
          conversationId,
          text,
          (event) => {
            if (event.kind === "delta") {
              setTurn("writing")
              draftRef.current += event.text
              setDraft(draftRef.current)
            } else if (event.kind === "message") {
              const message = event.message
              if (message.role === "user") {
                setMessages((prev) => {
                  const index = prev.findIndex((item) => item.id === local.id)
                  return index >= 0
                    ? prev.map((item, at) => (at === index ? { ...message } : item))
                    : [...prev, message]
                })
                setTurn("thinking")
              } else {
                setMessages((prev) => [...prev, message])
                clearDraft()
                setTurn("thinking")
                if (message.kind !== "text") scheduleRefresh()
              }
            } else if (event.kind === "tool") {
              setTurn("tool")
              scheduleRefresh()
            } else if (event.kind === "error") {
              setTurnError(event.error)
            }
          },
          controller.signal,
        )
      } catch (error) {
        if (!(error instanceof DOMException && error.name === "AbortError")) {
          setTurnError(error instanceof Error ? error.message : String(error))
        }
      } finally {
        abort.current = null
        // Whatever streamed but was not stored stays readable, marked as cut.
        const rest = draftRef.current
        if (rest) {
          setMessages((prev) => [
            ...prev,
            {
              id: `cut-${Date.now()}`,
              conversationId: conversationId!, 
              role: "assistant",
              kind: "text",
              text: rest,
              payload: '{"interrupted":true}',
              createdAt: new Date().toISOString(),
            },
          ])
          clearDraft()
        }
        setTurn("idle")
        scheduleRefresh()
      }
    },
    [activeId, clearDraft, hasKey, scheduleRefresh, turn],
  )

  const stop = React.useCallback(() => {
    if (activeId) api.stop(activeId).catch(() => {})
    abort.current?.abort()
    setTurn("idle")
  }, [activeId])

  // ---- answers, deletions, appearance -------------------------------------

  const answer = React.useCallback(async (messageId: string, question: number, choice: number) => {
    await api.answer(messageId, question, choice).catch(() => {})
    setMessages((prev) =>
      prev.map((message) => {
        if (message.id !== messageId) return message
        try {
          const payload = JSON.parse(message.payload)
          const item = payload.questions?.[question]
          if (item) item.chosen = choice
          return { ...message, payload: JSON.stringify(payload) }
        } catch {
          return message
        }
      }),
    )
    if (activeGoalId) scheduleRefresh()
  }, [activeGoalId, scheduleRefresh])

  const removeConversation = React.useCallback(
    async (id: string) => {
      await api.deleteConversation(id).catch(() => {})
      setConversations((prev) => prev.filter((item) => item.id !== id))
      if (id === activeId) {
        setActiveId(null)
        setMessages([])
        clearDraft()
      }
    },
    [activeId, clearDraft],
  )

  const removeRoute = React.useCallback(
    async (id: string) => {
      await api.deleteRoute(id).catch(() => {})
      setGoals((prev) => prev.filter((goal) => goal.id !== id))
      setConversations((prev) => prev.filter((item) => item.goalId !== id))
      if (id === activeGoalId) {
        setActiveId(null)
        setActiveGoalId(null)
        setMessages([])
        setRoute(null)
        setAbilities([])
      }
    },
    [activeGoalId],
  )

  const setAppearance = React.useCallback(async (next: Appearance) => {
    setAppearanceState(next)
    applyAppearance(next)
    await api.saveSettings({ appearance: next }).catch(() => {})
  }, [])

  React.useEffect(() => {
    return () => {
      if (refreshTimer.current) window.clearTimeout(refreshTimer.current)
      abort.current?.abort()
    }
  }, [])

  const view: "home" | "chat" = messages.length > 0 || turn !== "idle" || draft ? "chat" : "home"

  const value: CourseValue = {
    ready,
    bootError,
    goals,
    conversations,
    activeId,
    activeGoalId,
    messages,
    route,
    abilities,
    draft,
    turn,
    turnError,
    hasKey,
    appearance,
    model,
    view,
    busy: turn !== "idle",
    open,
    startNew,
    addRoute,
    submit,
    stop,
    answer,
    removeConversation,
    removeRoute,
    setAppearance,
    refresh,
  }

  return <CourseContext.Provider value={value}>{children}</CourseContext.Provider>
}
