import { useCourse } from "@/lib/course"
import { Reading } from "./markdown"
import { cn } from "@/lib/utils"
import type { QuizPayload } from "@/lib/types"

const LETTERS = ["A", "B", "C", "D"]

/** One question, answered by pressing the row. The answer is stored on the
 *  message, so it is still there after a reload. */
export function QuizCard({ messageId, payload }: { messageId: string; payload: QuizPayload }) {
  const { answer } = useCourse()
  const question = payload.questions?.[0]
  if (!question) return null

  const chosen = question.chosen
  const answered = typeof chosen === "number"
  const correct = answered && chosen === question.answerIndex

  return (
    <article className="pane my-9 rounded-[18px] px-7 pt-6 pb-7">
      <div className="flex items-center gap-2 text-[10px] tracking-[0.14em] text-brand">
        <span className="size-[3px] rounded-full bg-brand" />
        <span>小测</span>
      </div>

      <Reading text={question.prompt} className="reading-prompt mt-3 text-foreground" />

      <div className="mt-4 -mx-2 flex flex-col gap-[3px]">
        {question.options.map((option, index) => {
          const picked = chosen === index
          const isAnswer = question.answerIndex === index
          return (
            <button
              key={index}
              type="button"
              disabled={answered}
              onClick={() => void answer(messageId, 0, index)}
              className={cn(
                "flex w-full items-start gap-3 rounded-[10px] px-2.5 py-2.5 text-left transition-colors duration-150",
                "focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none",
                !answered && "hover:bg-secondary",
                answered && isAnswer && "bg-brand-muted",
                answered && picked && !isAnswer && "bg-destructive/10",
                answered && !picked && !isAnswer && "opacity-75",
              )}
            >
              <span
                className={cn(
                  "mt-[2px] w-4 shrink-0 text-[11px] tabular-nums",
                  answered && isAnswer ? "text-brand" : "text-muted-foreground",
                )}
              >
                {LETTERS[index]}
              </span>
              <span className="text-[14px] leading-[1.62] text-foreground/90">{option}</span>
            </button>
          )
        })}
      </div>

      {answered ? (
        <div className="mt-4 flex flex-col gap-1.5 border-t border-border/70 pt-4">
          <span className={cn("text-[11px] tracking-[0.04em]", correct ? "text-brand" : "text-foreground/80")}>
            {correct ? "答对了" : `答错了 · 正确答案是 ${LETTERS[question.answerIndex]}`}
          </span>
          {question.explanation ? (
            <p className="text-[13px] leading-[1.7] text-foreground/70">{question.explanation}</p>
          ) : null}
        </div>
      ) : null}
    </article>
  )
}
