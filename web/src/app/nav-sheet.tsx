import * as React from "react"
import { Plus, X } from "lucide-react"

import { useCourse } from "@/lib/course"
import { cn } from "@/lib/utils"
import { GlyphButton } from "./chrome"

/**
 * Routes and past conversations, on a sheet that lies on the desk. It is not
 * a modal: the page stays as it is behind it, and a click anywhere else puts
 * the sheet away.
 */
export function NavSheet({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const {
    goals,
    conversations,
    activeId,
    open: openConversation,
    startNew,
    addRoute,
    removeConversation,
    removeRoute,
  } = useCourse()

  const sheet = React.useRef<HTMLDivElement>(null)
  const close = () => onOpenChange(false)

  React.useEffect(() => {
    if (!open) return
    sheet.current?.focus()

    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") onOpenChange(false)
    }
    const onDown = (event: MouseEvent) => {
      if (sheet.current && !sheet.current.contains(event.target as Node)) onOpenChange(false)
    }

    // The click that opened the sheet has already happened; listen from the
    // next tick so it cannot close what it just opened.
    const timer = window.setTimeout(() => document.addEventListener("mousedown", onDown), 0)
    document.addEventListener("keydown", onKey)
    return () => {
      window.clearTimeout(timer)
      document.removeEventListener("mousedown", onDown)
      document.removeEventListener("keydown", onKey)
    }
  }, [open, onOpenChange])

  if (!open) return null

  return (
    <div
      ref={sheet}
      role="dialog"
      aria-label="路线与对话"
      tabIndex={-1}
      className="fixed top-4 left-4 z-40 flex max-h-[calc(100dvh-2rem)] w-[318px] flex-col rounded-[16px] bg-popover text-popover-foreground outline-none data-open:animate-in data-open:fade-in-0 data-open:slide-in-from-left-2 data-open:duration-150"
      data-open=""
    >
      <header className="flex h-11 shrink-0 items-center justify-between pr-2 pl-4">
        <span className="text-[11.5px] tracking-[0.04em] text-muted-foreground">路线与对话</span>
        <div className="flex items-center gap-0.5">
          <GlyphButton
            label="新对话"
            className="size-7"
            onClick={() => {
              void startNew()
              close()
            }}
          >
            <Plus className="size-3.5" strokeWidth={1.6} />
          </GlyphButton>
          <GlyphButton label="关闭" className="size-7" onClick={close}>
            <X className="size-3.5" strokeWidth={1.6} />
          </GlyphButton>
        </div>
      </header>

      <div className="h-px shrink-0 bg-border/70" />

      <div className="min-h-0 flex-1 overflow-y-auto px-2 py-2">
        {goals.length === 0 ? (
          <p className="px-2.5 py-3 text-[11px] leading-[1.7] text-muted-foreground">
            还没有路线。问一个问题，路线会随对话长出来。
          </p>
        ) : (
          goals.map((goal) => {
            const own = conversations.filter((conversation) => conversation.goalId === goal.id)
            return (
              <section key={goal.id} className="mb-3">
                <header className="group/route flex items-center gap-2 px-2.5 pt-2 pb-1.5">
                  <span className="size-[3px] shrink-0 rounded-full bg-brand/80" />
                  <span className="min-w-0 flex-1 truncate font-serif text-[12.5px] tracking-[0.02em] text-foreground">
                    {goal.title}
                  </span>
                  <span className="shrink-0 font-mono text-[9px] tabular-nums text-muted-foreground">
                    {goal.done}/{goal.total}
                  </span>
                  <button
                    type="button"
                    aria-label={`删除路线 ${goal.title}`}
                    onClick={() => {
                      if (window.confirm(`删除《${goal.title}》和它的全部对话？`)) void removeRoute(goal.id)
                    }}
                    className="shrink-0 text-transparent transition-colors group-hover/route:text-muted-foreground/60 hover:text-destructive focus-visible:text-destructive focus-visible:outline-none"
                  >
                    <X className="size-3" strokeWidth={1.6} />
                  </button>
                </header>

                <div className="ml-[13px] border-l border-border/70 pl-1.5">
                  {own.map((conversation) => {
                    const active = activeId === conversation.id
                    return (
                      <div
                        key={conversation.id}
                        className={cn(
                          "group/row relative flex items-center rounded-[8px] transition-colors",
                          active ? "bg-secondary" : "hover:bg-secondary/60",
                        )}
                      >
                        {active ? (
                          <span className="absolute top-1/2 -left-[7px] h-3 w-[2px] -translate-y-1/2 rounded-full bg-brand" />
                        ) : null}
                        <button
                          type="button"
                          onClick={() => {
                            void openConversation(conversation.id)
                            close()
                          }}
                          className="min-w-0 flex-1 py-1.5 pr-1 pl-2 text-left focus-visible:outline-none"
                        >
                          <span
                            className={cn(
                              "block truncate text-[11.5px] leading-[1.6]",
                              active ? "text-foreground" : "text-muted-foreground",
                            )}
                          >
                            {conversation.title || "新的对话"}
                          </span>
                        </button>
                        <button
                          type="button"
                          aria-label="删除对话"
                          onClick={() => void removeConversation(conversation.id)}
                          className="mr-1.5 shrink-0 text-transparent transition-colors group-hover/row:text-muted-foreground/50 hover:text-destructive focus-visible:text-destructive focus-visible:outline-none"
                        >
                          <X className="size-3" strokeWidth={1.6} />
                        </button>
                      </div>
                    )
                  })}

                  <button
                    type="button"
                    onClick={() => {
                      void startNew(goal.id)
                      close()
                    }}
                    className="mt-0.5 w-full rounded-[8px] py-1.5 pl-2 text-left text-[11px] text-muted-foreground transition-colors hover:bg-secondary/50 hover:text-foreground focus-visible:outline-none"
                  >
                    + 新对话
                  </button>
                </div>
              </section>
            )
          })
        )}
      </div>

      <div className="h-px shrink-0 bg-border/70" />

      <footer className="flex shrink-0 items-center justify-between px-3 py-2.5">
        <button
          type="button"
          onClick={() => {
            void addRoute("新的学习路线")
            close()
          }}
          className="text-[11px] text-muted-foreground transition-colors hover:text-foreground focus-visible:text-foreground focus-visible:outline-none"
        >
          新建路线
        </button>
        <span className="font-mono text-[9px] text-muted-foreground/70">{conversations.length} 段对话</span>
      </footer>
    </div>
  )
}
