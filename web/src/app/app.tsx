import * as React from "react"
import { Settings2 } from "lucide-react"

import { CourseProvider, useCourse } from "@/lib/course"
import { AbilityRail } from "./ability-rail"
import { GlyphButton, Mark } from "./chrome"
import { ComposerBar } from "./composer-bar"
import { Home } from "./home"
import { NavSheet } from "./nav-sheet"
import { RouteRail } from "./route-rail"
import { SettingsDialog } from "./settings-dialog"
import { Transcript, useFollow } from "./transcript"

export function App() {
  return (
    <CourseProvider>
      <Desk />
    </CourseProvider>
  )
}

function Desk() {
  const { ready, bootError, view } = useCourse()
  const [nav, setNav] = React.useState(false)
  const [settings, setSettings] = React.useState(false)

  if (!ready) return <div className="min-h-dvh" />

  if (bootError) {
    return (
      <div className="flex min-h-dvh items-center justify-center px-6">
        <div className="max-w-[380px] text-center">
          <p className="text-[12px] leading-relaxed text-muted-foreground">连不上本地服务。</p>
          <p className="mt-1.5 font-mono text-[10.5px] leading-relaxed text-muted-foreground/60">{bootError}</p>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="mt-4 text-[11px] text-brand underline decoration-brand/30 underline-offset-4"
          >
            再试一次
          </button>
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="pointer-events-none fixed inset-x-0 top-0 z-20 h-16 bg-gradient-to-b from-background via-background/85 to-transparent" />
      <header className="fixed inset-x-0 top-0 z-30 flex h-14 items-center justify-between px-3">
        <GlyphButton label="路线与对话" onClick={() => setNav(true)}>
          <Mark className="size-[13px] text-brand" />
        </GlyphButton>
        <GlyphButton label="设置" onClick={() => setSettings(true)}>
          <Settings2 className="size-[15px]" strokeWidth={1.6} />
        </GlyphButton>
      </header>

      {view === "home" ? <Home /> : <Chat />}

      <NavSheet open={nav} onOpenChange={setNav} />
      <SettingsDialog open={settings} onOpenChange={setSettings} />
    </>
  )
}

/** A conversation: one column of reading, held between two margins. */
function Chat() {
  const end = useFollow(true)
  const { activeId } = useCourse()

  // Opening a conversation lands at its foot, where the talk continues.
  React.useEffect(() => {
    requestAnimationFrame(() => {
      window.scrollTo({ top: document.documentElement.scrollHeight })
    })
  }, [activeId])

  return (
    <>
      <RouteRail />
      <AbilityRail />

      <main className="mx-auto w-full max-w-[680px] px-6 pt-20 pb-[248px] duration-300 animate-in fade-in-0 slide-in-from-bottom-1 lg:px-0">
        <Transcript />
        <div ref={end} className="h-px" />
      </main>

      {/* The foot of the page: a soft fade, then solid paper behind the box
          that holds the next line, so nothing reads through it. */}
      <div className="pointer-events-none fixed inset-x-0 bottom-0 z-10">
        <div className="h-[54px] bg-gradient-to-t from-background to-transparent" />
        <div className="h-[108px] bg-background" />
      </div>

      <div className="fixed inset-x-0 bottom-0 z-20 pb-4">
        <div className="mx-auto w-full max-w-[680px] px-6 lg:px-0">
          <ComposerBar />
        </div>
      </div>
    </>
  )
}
