import * as React from "react"

import { cn } from "@/lib/utils"
import type { CanvasBlock, CanvasPayload } from "@/lib/types"
import { Reading } from "./markdown"

/** The controlled blocks the tutor may draw on the canvas. No HTML, no script:
 *  each type below is a component of ours. */
export function Canvas({ payload }: { payload: CanvasPayload }) {
  const blocks = payload.blocks ?? []
  if (blocks.length === 0) return null

  return (
    <section className="my-8 flex flex-col gap-5">
      {payload.title ? (
        <h3 className="font-serif text-[17px] leading-snug text-foreground">{payload.title}</h3>
      ) : null}
      {blocks.map((block, index) => (
        <Block key={index} block={block} />
      ))}
    </section>
  )
}

function Block({ block }: { block: CanvasBlock }) {
  switch (block.type) {
    case "text":
      return <Reading text={block.text ?? ""} />

    case "quote":
      return (
        <figure className="border-l border-brand/55 pl-4">
          <blockquote className="text-[14px] leading-[1.8] text-foreground/75">{block.text}</blockquote>
          {block.caption ? (
            <figcaption className="mt-2 text-[10.5px] text-muted-foreground/80">{block.caption}</figcaption>
          ) : null}
        </figure>
      )

    case "callout": {
      const tone =
        block.tone === "success"
          ? "border-success/60"
          : block.tone === "warning"
            ? "border-status-interrupted/60"
            : "border-brand/55"
      const label =
        block.tone === "success" ? "text-success" : block.tone === "warning" ? "text-status-interrupted" : "text-brand"
      return (
        <div className={cn("border-l pl-4", tone)}>
          {block.title ? <div className={cn("text-[10px] tracking-[0.12em]", label)}>{block.title}</div> : null}
          {block.body ? <p className="mt-1 text-[13.5px] leading-[1.75] text-foreground/85">{block.body}</p> : null}
        </div>
      )
    }

    case "steps":
      return (
        <ol className="flex flex-col gap-3">
          {(block.items ?? []).map((item, index) => (
            <li key={index} className="flex gap-3">
              <span className="mt-[2px] w-4 shrink-0 font-mono text-[10px] tabular-nums text-brand/80">
                {String(index + 1).padStart(2, "0")}
              </span>
              <span className="min-w-0">
                {item.title ? (
                  <span className="block text-[13.5px] leading-[1.6] text-foreground">{item.title}</span>
                ) : null}
                {item.body ? (
                  <span className="mt-0.5 block text-[12.5px] leading-[1.7] text-muted-foreground">{item.body}</span>
                ) : null}
              </span>
            </li>
          ))}
        </ol>
      )

    case "table":
      return (
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left text-[12.5px] leading-[1.6]">
            <thead>
              <tr>
                {(block.columns ?? []).map((column, index) => (
                  <th
                    key={index}
                    className="border-b border-border pb-2 pr-4 text-[10.5px] tracking-[0.04em] text-muted-foreground"
                  >
                    {column}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {(block.rows ?? []).map((row, rowIndex) => (
                <tr key={rowIndex}>
                  {row.map((cell, cellIndex) => (
                    <td
                      key={cellIndex}
                      className="border-b border-border/55 py-2 pr-4 align-top text-foreground/85 last:border-b-0"
                    >
                      {cell}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )

    case "compare":
      return (
        <div className="grid gap-6 sm:grid-cols-2">
          {[block.left, block.right].map((side, index) =>
            side ? (
              <div key={index} className="min-w-0">
                <div className="text-[10px] tracking-[0.12em] text-muted-foreground">{side.title}</div>
                <ul className="mt-2 flex flex-col gap-1.5">
                  {(side.items ?? []).map((item, at) => (
                    <li key={at} className="flex gap-2 text-[13px] leading-[1.7] text-foreground/80">
                      <span className="mt-[9px] size-[3px] shrink-0 rounded-full bg-brand/60" />
                      <span>{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ) : null,
          )}
        </div>
      )

    case "code":
      return (
        <div>
          {block.language ? (
            <div className="mb-1.5 font-mono text-[9.5px] tracking-[0.1em] text-muted-foreground/70">
              {block.language}
            </div>
          ) : null}
          <pre className="overflow-x-auto rounded-[10px] bg-secondary/60 p-3.5 font-mono text-[12px] leading-[1.65] text-foreground/85 ring-1 ring-border">
            <code>{block.code}</code>
          </pre>
          {block.caption ? (
            <div className="mt-2 text-[10.5px] text-muted-foreground/80">{block.caption}</div>
          ) : null}
        </div>
      )

    case "flashcards":
      return <Flashcards cards={block.cards ?? []} />

    case "bar":
      return <Bars labels={block.labels ?? []} values={block.values ?? []} unit={block.unit} title={block.title} />

    case "divider":
      return <hr className="border-0 border-t border-border/70" />

    default:
      return null
  }
}

function Flashcards({ cards }: { cards: { front?: string; back?: string }[] }) {
  const [flipped, setFlipped] = React.useState<number[]>([])
  return (
    <div className="grid gap-2 sm:grid-cols-2">
      {cards.map((card, index) => {
        const open = flipped.includes(index)
        return (
          <button
            key={index}
            type="button"
            onClick={() =>
              setFlipped((prev) => (prev.includes(index) ? prev.filter((item) => item !== index) : [...prev, index]))
            }
            className="group flex min-h-[86px] flex-col justify-between gap-3 rounded-[10px] bg-secondary/50 p-3.5 text-left ring-1 ring-border transition-colors duration-150 hover:bg-secondary focus-visible:ring-ring focus-visible:outline-none"
          >
            <span className="text-[10px] tracking-[0.12em] text-muted-foreground/60">
              {open ? "背面" : "正面"}
            </span>
            <span className={cn("text-[13px] leading-[1.6]", open ? "text-brand" : "text-foreground/90")}>
              {open ? card.back : card.front}
            </span>
          </button>
        )
      })}
    </div>
  )
}

function Bars({
  labels,
  values,
  unit,
  title,
}: {
  labels: string[]
  values: number[]
  unit?: string
  title?: string
}) {
  const max = Math.max(...values.map((value) => Math.abs(value)), 0) || 1
  return (
    <div className="flex flex-col gap-2">
      {title ? <div className="text-[10px] tracking-[0.12em] text-muted-foreground">{title}</div> : null}
      {labels.map((label, index) => {
        const value = values[index] ?? 0
        return (
          <div key={index} className="flex items-center gap-3">
            <span className="w-20 shrink-0 truncate text-[11px] text-muted-foreground">{label}</span>
            <span className="h-[3px] flex-1 overflow-hidden rounded-full bg-secondary">
              <span
                className="block h-full rounded-full bg-brand/75"
                style={{ width: `${Math.max(2, (Math.abs(value) / max) * 100)}%` }}
              />
            </span>
            <span className="w-12 shrink-0 text-right font-mono text-[10.5px] tabular-nums text-muted-foreground">
              {value}
              {unit ?? ""}
            </span>
          </div>
        )
      })}
    </div>
  )
}
