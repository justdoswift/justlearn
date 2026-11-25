import * as React from "react"

import type { LessonPayload } from "@/lib/types"
import { Reading } from "./markdown"

/**
 * A lesson the tutor set down for this node: a plate on the page, not a chat
 * card. Serif title, the body as prose, and the few things worth keeping.
 */
export function LessonCard({ payload }: { payload: LessonPayload }) {
  const [revealed, setRevealed] = React.useState(false)
  const keyPoints = payload.keyPoints ?? []

  return (
    <article className="pane my-9 rounded-[18px] px-7 pt-6 pb-7">
      {payload.kicker ? (
        <div className="flex items-center gap-2 text-[10px] tracking-[0.14em] text-brand">
          <span className="size-[3px] rounded-full bg-brand" />
          <span className="truncate">{payload.kicker}</span>
        </div>
      ) : null}

      <h2 className="mt-3 font-serif text-[21px] leading-[1.36] tracking-[0.01em] text-foreground">
        {payload.title}
      </h2>

      <Reading text={payload.body} className="mt-4" />

      {payload.code ? (
        <pre className="mt-5 overflow-x-auto rounded-[10px] p-3.5 font-mono text-[12px] leading-[1.65] ring-1 ring-border">
          <code>{payload.code}</code>
        </pre>
      ) : null}

      {keyPoints.length > 0 ? (
        <ul className="mt-6 space-y-2 border-t border-border/70 pt-5">
          {keyPoints.map((point) => (
            <li key={point} className="flex gap-2.5 text-[12.5px] leading-[1.7] text-foreground/80">
              <span className="mt-[9px] size-[3px] shrink-0 rounded-full bg-brand/70" />
              <span>{point}</span>
            </li>
          ))}
        </ul>
      ) : null}

      {payload.mnemonic ? (
        <p className="mt-5 text-[12px] leading-relaxed text-muted-foreground">
          <span className="mr-1.5 text-[10px] tracking-[0.14em] text-brand">记</span>
          {payload.mnemonic}
        </p>
      ) : null}

      {payload.recall?.prompt ? (
        <div className="mt-6 border-l border-brand/50 pl-4">
          <div className="text-[10px] tracking-[0.14em] text-brand/80">回想</div>
          <p className="mt-1.5 text-[13.5px] leading-[1.7] text-foreground/90">{payload.recall.prompt}</p>
          {revealed ? (
            <p className="mt-2 text-[13px] leading-[1.7] text-foreground/70">{payload.recall.answer}</p>
          ) : (
            <button
              type="button"
              onClick={() => setRevealed(true)}
              className="mt-2 text-[11px] text-muted-foreground underline decoration-muted-foreground/30 underline-offset-4 transition-colors hover:text-foreground hover:decoration-foreground/50 focus-visible:text-foreground focus-visible:outline-none"
            >
              先在心里答一遍，再看答案
            </button>
          )}
        </div>
      ) : null}
    </article>
  )
}
