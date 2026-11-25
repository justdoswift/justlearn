import * as React from "react"

import { useCourse } from "@/lib/course"
import { cn } from "@/lib/utils"
import type { Node } from "@/lib/types"

type Stage = { title: string; nodes: Node[] }

const SCALES = [1, 0.7, 0.5, 0.36]

/** The roadmap as a scale in the left margin: one tick per stage, no panel,
 *  no labels on the page. A tick opens its stage on the side. */
export function RouteRail() {
  const { route } = useCourse()
  const [open, setOpen] = React.useState<{ index: number; top: number } | null>(null)
  const [pinned, setPinned] = React.useState<number | null>(null)

  const stages: Stage[] = []
  for (const node of route?.nodes ?? []) {
    const title = node.stageTitle || "阶段"
    let stage = stages.find((item) => item.title === title)
    if (!stage) {
      stage = { title, nodes: [] }
      stages.push(stage)
    }
    stage.nodes.push(node)
  }

  if (stages.length === 0) return null

  const currentStage = Math.max(
    0,
    stages.findIndex((stage) => stage.title === route?.current?.stageTitle),
  )
  const active = open ? open.index : pinned
  const shown = active !== null ? stages[active] : null

  const show = (index: number, element: HTMLElement) => {
    const rect = element.getBoundingClientRect()
    const height = 240
    const top = Math.min(
      Math.max(rect.top + rect.height / 2, 16 + height / 2),
      window.innerHeight - 16 - height / 2,
    )
    setOpen({ index, top })
  }

  return (
    <>
      <nav
        aria-label="路线"
        className="fixed top-1/2 left-5 z-20 hidden -translate-y-1/2 flex-col lg:flex"
        onMouseLeave={() => setOpen(null)}
      >
        {stages.map((stage, index) => {
          const distance = Math.abs(index - currentStage)
          const scale = SCALES[Math.min(distance, SCALES.length - 1)]
          const isCurrent = index === currentStage
          const done = stage.nodes.every((node) => node.status === "done")
          return (
            <button
              key={stage.title + index}
              type="button"
              aria-label={`${stage.title}，${stage.nodes.length} 节`}
              onMouseEnter={(event) => show(index, event.currentTarget)}
              onFocus={(event) => show(index, event.currentTarget)}
              onBlur={() => setOpen(null)}
              onClick={() => setPinned(pinned === index ? null : index)}
              className="group flex h-[24px] w-[28px] items-center focus-visible:outline-none"
            >
              <span
                className={cn(
                  "block h-[2px] rounded-full transition-all duration-300 ease-out",
                  isCurrent
                    ? "bg-brand"
                    : done
                      ? "bg-brand/60"
                      : "bg-foreground/45 group-hover:bg-foreground/65",
                  open?.index === index && !isCurrent && "bg-foreground/65",
                )}
                style={{ width: 22 * scale }}
              />
            </button>
          )
        })}
      </nav>

      {shown && open ? (
        <div
          className="pointer-events-none fixed left-[54px] z-20 hidden w-[280px] -translate-y-1/2 lg:block"
          style={{ top: open.top }}
        >
          <div className="pane rounded-[14px] px-4 py-3.5">
            <div className="flex items-baseline justify-between gap-3">
              <h3 className="truncate text-[13.5px] text-foreground">{shown.title}</h3>
              <span className="shrink-0 font-mono text-[9px] tracking-[0.08em] text-muted-foreground/70">
                {String((open.index + 1)).padStart(2, "0")} / {String(stages.length).padStart(2, "0")}
              </span>
            </div>

            <div className="mt-2.5 h-[2px] w-full overflow-hidden rounded-full bg-foreground/10">
              <div
                className="h-full rounded-full bg-brand/80"
                style={{
                  width: `${Math.round(
                    (shown.nodes.filter((node) => node.status === "done").length / shown.nodes.length) * 100,
                  )}%`,
                }}
              />
            </div>

            <ul className="mt-3 flex flex-col gap-1.5">
              {shown.nodes.slice(0, 5).map((node) => {
                const now = route?.current?.id === node.id
                return (
                  <li key={node.id} className="flex items-baseline gap-2">
                    <span
                      className={cn(
                        "mt-[1px] size-[3px] shrink-0 rounded-full",
                        now ? "bg-brand" : node.status === "done" ? "bg-brand/55" : "bg-foreground/20",
                      )}
                    />
                    <span
                      className={cn(
                        "min-w-0 flex-1 truncate text-[11.5px] leading-[1.6]",
                        now ? "text-foreground" : "text-muted-foreground",
                      )}
                    >
                      {node.title}
                    </span>
                    {now ? <span className="shrink-0 text-[9px] text-brand">现在</span> : null}
                  </li>
                )
              })}
              {shown.nodes.length > 5 ? (
                <li className="pl-[11px] text-[10px] text-muted-foreground/60">还有 {shown.nodes.length - 5} 节</li>
              ) : null}
            </ul>
          </div>
        </div>
      ) : null}
    </>
  )
}
