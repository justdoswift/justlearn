export type Goal = {
  id: string
  title: string
  level: string
  done: number
  total: number
  percent: number
  currentTitle: string
}

export type Conversation = {
  id: string
  goalId: string
  title: string
  subtitle: string
  updatedAt: string
}

export type Cell = {
  id: string
  nodeId: string
  abilityId: string
  title: string
  description: string
  mastery: number
  evidence: number
  lastEvidence: string
}

export type Node = {
  id: string
  goalId: string
  seq: number
  stageTitle: string
  title: string
  summary: string
  status: string
  threshold: number
  cells: Cell[]
}

export type Ability = {
  id: string
  goalId: string
  seq: number
  title: string
  description: string
}

export type Message = {
  id: string
  conversationId: string
  role: "user" | "assistant"
  kind: string
  text: string
  payload: string
  createdAt: string
}

export type Route = {
  goal: Goal
  nodes: Node[]
  current: Node
}

export type Bootstrap = {
  goals: Goal[]
  conversations: Conversation[]
  activeId: string
  hasKey: boolean
  model: string
  appearance: Appearance
}

export type Appearance = "light" | "dark" | "system"

export type Settings = {
  baseUrl: string
  model: string
  appearance: Appearance
  hasKey: boolean
}

export type TurnEvent =
  | { kind: "delta"; text: string }
  | { kind: "message"; message: Message }
  | { kind: "tool" }
  | { kind: "error"; error: string }
  | { kind: "done" }

export type LessonPayload = {
  nodeId?: string
  kicker?: string
  title: string
  body: string
  recall?: { prompt: string; answer: string }
  code?: string
  mnemonic?: string
  keyPoints?: string[]
}

export type QuizQuestion = {
  id: string
  prompt: string
  options: string[]
  answerIndex: number
  explanation?: string
  chosen?: number
}

export type QuizPayload = {
  nodeId?: string
  questions: QuizQuestion[]
}

export type CanvasBlock = {
  type: string
  text?: string
  tone?: string
  title?: string
  body?: string
  caption?: string
  columns?: string[]
  rows?: string[][]
  items?: { title?: string; body?: string }[]
  left?: { title?: string; items?: string[] }
  right?: { title?: string; items?: string[] }
  language?: string
  code?: string
  cards?: { front?: string; back?: string }[]
  labels?: string[]
  values?: number[]
  unit?: string
}

export type CanvasPayload = {
  nodeId?: string
  title?: string
  blocks: CanvasBlock[]
}
