import { Mark, Wordmark } from "./chrome"
import { ComposerBar } from "./composer-bar"
import { useCourse } from "@/lib/course"

/**
 * The desk before anything is on it: the mark, and the box you write in.
 * Nothing else — no suggestions, no history list, no counts.
 */
export function Home() {
  const { conversations, open } = useCourse()
  const last = conversations.find((item) => item.title && item.title !== "新对话")

  return (
    <div className="flex min-h-dvh flex-col items-center justify-center px-6 pb-[6vh]">
      <div className="w-full max-w-[680px]">
        <div className="mb-8 flex items-center justify-center gap-3">
          <Mark className="size-[12px] text-brand" />
          <Wordmark />
        </div>
        <ComposerBar hero />
      </div>

      {last ? (
        <button
          type="button"
          onClick={() => void open(last.id)}
          className="group mt-7 flex max-w-[440px] items-center gap-1.5 text-[12px] text-muted-foreground transition-colors duration-150 hover:text-foreground focus-visible:text-foreground focus-visible:outline-none"
        >
          <span>继续</span>
          <span className="truncate font-serif text-[13px] text-foreground/80 group-hover:text-foreground">
            {last.title || "上一次"}
          </span>
          <span className="translate-y-[-0.5px] text-muted-foreground/70 transition-transform duration-150 group-hover:translate-x-0.5">
            →
          </span>
        </button>
      ) : null}
    </div>
  )
}
