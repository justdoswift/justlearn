import * as React from "react"

import { Composer, ComposerFooter } from "@/components/ui/composer"
import { useCourse } from "@/lib/course"
import { cn } from "@/lib/utils"

/**
 * The one way into a conversation. On the home page it is the whole page; in
 * a conversation it stands on the paper at the foot of the reading column.
 */
export function ComposerBar({ hero = false }: { hero?: boolean }) {
  const { submit, stop, busy, model, turnError, route } = useCourse()
  const [value, setValue] = React.useState("")

  const send = (text: string) => {
    if (!text.trim() || busy) return
    setValue("")
    void submit(text)
  }

  return (
    <div className="w-full">
      <Composer
        value={value}
        onValueChange={setValue}
        onSubmit={send}
        onStop={stop}
        running={busy}
        placeholder={hero ? "想弄懂点什么？" : "继续说，或者问我一个问题"}
        sendLabel="发送"
        stopLabel="停止"
        className={cn(
          "jl-composer rounded-[20px] transition-shadow duration-200",
          hero ? "px-2.5 py-3" : "px-2 py-2.5",
        )}
      />
      <ComposerFooter className={cn("mt-2 px-0", hero && "justify-center")}>
        {turnError ? (
          <span className="text-[11px] leading-snug text-destructive/90">{turnError}</span>
        ) : hero ? null : (
          <span className="truncate text-[11px] tracking-[0.02em] text-foreground/70">
            {busy
              ? "生成中，可随时停止"
              : route?.current?.title
                ? `第 ${route.current.seq} / ${route.nodes.length} 节 · ${route.current.title}`
                : model}
          </span>
        )}
      </ComposerFooter>
    </div>
  )
}
