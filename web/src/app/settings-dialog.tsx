import * as React from "react"
import { X } from "lucide-react"

import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { api } from "@/lib/api"
import { useCourse } from "@/lib/course"
import { cn } from "@/lib/utils"
import { GlyphButton } from "./chrome"

const APPEARANCES = [
  { value: "system", label: "跟随系统" },
  { value: "light", label: "浅色" },
  { value: "dark", label: "深色" },
] as const

/** Three things only: how the desk looks, which model answers, and what this
 *  build is. The key never reaches the browser. */
export function SettingsDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { appearance, setAppearance } = useCourse()
  const [baseUrl, setBaseUrl] = React.useState("")
  const [model, setModel] = React.useState("")
  const [hasKey, setHasKey] = React.useState(true)
  const [state, setState] = React.useState<"idle" | "saving" | "saved" | "error">("idle")

  React.useEffect(() => {
    if (!open) return
    setState("idle")
    api
      .settings()
      .then((settings) => {
        setBaseUrl(settings.baseUrl)
        setModel(settings.model)
        setHasKey(settings.hasKey)
      })
      .catch(() => setState("error"))
  }, [open])

  const save = async () => {
    setState("saving")
    try {
      await api.saveSettings({ baseUrl: baseUrl.trim(), model: model.trim() })
      setState("saved")
    } catch {
      setState("error")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className="w-[460px] max-w-[calc(100vw-2rem)] gap-0 rounded-[16px] p-0 ring-0 shadow-[inset_0_1px_0_var(--glass-rim),0_0_0_1px_var(--popover-edge),0_40px_80px_-32px_var(--glass-shade),0_8px_20px_-8px_var(--glass-lift)]"
      >
        <DialogTitle className="sr-only">设置</DialogTitle>

        <header className="flex h-11 items-center justify-between pr-2 pl-4">
          <span className="text-[11.5px] tracking-[0.04em] text-muted-foreground">设置</span>
          <GlyphButton label="关闭" className="size-7" onClick={() => onOpenChange(false)}>
            <X className="size-3.5" strokeWidth={1.6} />
          </GlyphButton>
        </header>

        <div className="h-px bg-border/70" />

        <div className="px-4 py-4">
          <Field label="外观">
            <div className="mt-2 flex gap-1 rounded-[9px] bg-secondary/70 p-1">
              {APPEARANCES.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => void setAppearance(option.value)}
                  className={cn(
                    "flex-1 rounded-[7px] py-1.5 text-[11px] transition-colors duration-150 focus-visible:outline-none",
                    appearance === option.value
                      ? "bg-card text-foreground shadow-[inset_0_1px_0_var(--glass-rim),0_0_0_1px_var(--glass-edge)]"
                      : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </Field>

          <div className="my-4 h-px bg-border/70" />

          <Field label="模型">
            <div className="mt-2 flex flex-col gap-3">
              <div>
                <label htmlFor="settings-base-url" className="text-[10.5px] text-muted-foreground">
                  接口地址
                </label>
                <Input
                  id="settings-base-url"
                  value={baseUrl}
                  onChange={(event) => setBaseUrl(event.target.value)}
                  placeholder="https://api.openai.com/v1"
                  className="mt-1 h-8 text-[12px]"
                  spellCheck={false}
                />
              </div>
              <div>
                <label htmlFor="settings-model" className="text-[10.5px] text-muted-foreground">
                  模型名称
                </label>
                <Input
                  id="settings-model"
                  value={model}
                  onChange={(event) => setModel(event.target.value)}
                  placeholder="留空则用 .env 里的 PI_MODEL"
                  className="mt-1 h-8 text-[12px]"
                  spellCheck={false}
                />
              </div>
            </div>
            <div className="mt-2 flex items-center gap-1.5 text-[10px] text-muted-foreground/70">
              <span className={cn("size-[3px] rounded-full", hasKey ? "bg-brand" : "bg-status-interrupted")} />
              {hasKey ? "密钥已配置，只存在服务端" : "没有密钥：在 .env 里填写 PI_API_KEY"}
            </div>
          </Field>

          <div className="my-4 h-px bg-border/70" />

          <div className="flex items-center justify-between">
            <div>
              <div className="text-[11px] text-foreground/80">JustLearn</div>
              <div className="mt-0.5 text-[10px] text-muted-foreground/60">网页版 · 本地运行</div>
            </div>
            <button
              type="button"
              onClick={() => void save()}
              disabled={state === "saving"}
              className={cn(
                "h-7 rounded-[8px] px-2.5 text-[12px] transition-colors duration-150 focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none",
                "text-brand hover:bg-brand-muted disabled:opacity-60",
              )}
            >
              {state === "saving" ? "保存中" : state === "saved" ? "已保存" : state === "error" ? "重试" : "保存"}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="text-[10.5px] tracking-[0.06em] text-muted-foreground">{label}</div>
      {children}
    </div>
  )
}
