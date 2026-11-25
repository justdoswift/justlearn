import * as React from "react"

import { useCourse } from "@/lib/course"
import { StatusGlyph } from "@/components/ui/status-glyph"
import type { CanvasPayload, LessonPayload, Message, QuizPayload } from "@/lib/types"
import { Canvas } from "./canvas"
import { LessonCard } from "./lesson-card"
import { QuizCard } from "./quiz-card"
import { Reading } from "./markdown"

function parse<T>(payload: string): T | null {
  if (!payload) return null
  try {
    return JSON.parse(payload) as T
  } catch {
    return null
  }
}

/** The conversation, read top to bottom: the reader's question as a line of
 *  ink with a jade mark, and the tutor's answer as a page. */
export function Transcript() {
  const { messages, draft, turn } = useCourse()

  return (
    <div className="flex flex-col">
      {messages.map((message) => (
        <Row key={message.id} message={message} />
      ))}

      {draft ? (
        <div className="mb-10">
          <Reading text={draft} />
          <span className="caret" />
        </div>
      ) : null}

      {turn === "thinking" ? (
        <div className="mb-10 flex items-center gap-2 text-[11px] text-muted-foreground/80">
          <StatusGlyph activity="thinking" className="size-3" />
          <span>在想</span>
        </div>
      ) : null}

      {turn === "tool" ? (
        <div className="mb-10 flex items-center gap-2 text-[11px] text-muted-foreground/80">
          <StatusGlyph activity="running" className="size-3" />
          <span>正在调整路线</span>
        </div>
      ) : null}
    </div>
  )
}

function Row({ message }: { message: Message }) {
  if (message.role === "user") {
    return (
      <div className="mt-2 mb-10 border-l-[3px] border-brand/70 pl-4">
        <p className="max-w-[560px] font-serif text-[15px] leading-[1.6] tracking-[0.015em] whitespace-pre-wrap text-foreground/85">
          {message.text}
        </p>
      </div>
    )
  }

  if (message.kind === "lesson") {
    const payload = parse<LessonPayload>(message.payload)
    return payload ? <LessonCard payload={payload} /> : <Text message={message} />
  }

  if (message.kind === "quiz") {
    const payload = parse<QuizPayload>(message.payload)
    return payload ? <QuizCard messageId={message.id} payload={payload} /> : <Text message={message} />
  }

  if (message.kind === "ui") {
    const payload = parse<CanvasPayload>(message.payload)
    return payload ? <Canvas payload={payload} /> : <Text message={message} />
  }

  return <Text message={message} />
}

function Text({ message }: { message: Message }) {
  const cut = parse<{ interrupted?: boolean }>(message.payload)?.interrupted === true
  if (!message.text.trim()) return null
  return (
    <div className="mb-10">
      <Reading text={message.text} />
      {cut ? (
        <p className="mt-3 text-[10.5px] tracking-[0.03em] text-status-interrupted/80">这一轮中断了，上面是已经写完的部分。</p>
      ) : null}
    </div>
  )
}

/** Keeps the newest line in view while the tutor writes. A message the reader
 *  just sent always comes into view; anything else only when the reader is
 *  already near the foot of the page, so reading back is never interrupted. */
export function useFollow(enabled: boolean) {
  const ref = React.useRef<HTMLDivElement>(null)
  const { messages, draft } = useCourse()
  const count = React.useRef(messages.length)

  const nearFoot = () =>
    document.documentElement.scrollHeight - window.scrollY - window.innerHeight < 320
  const toFoot = () =>
    window.scrollTo({ top: document.documentElement.scrollHeight, behavior: "smooth" })

  React.useEffect(() => {
    if (!enabled) return
    const grew = messages.length > count.current
    const last = messages[messages.length - 1]
    count.current = messages.length
    if (grew && last?.role === "user") {
      toFoot()
      return
    }
    if (nearFoot()) toFoot()
  }, [enabled, messages])

  React.useEffect(() => {
    if (!enabled || !draft) return
    if (nearFoot()) toFoot()
  }, [enabled, draft])

  return ref
}
