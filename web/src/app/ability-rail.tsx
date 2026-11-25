import * as React from "react"

import { useCourse } from "@/lib/course"
import { cn } from "@/lib/utils"
import type { Ability, Cell } from "@/lib/types"

const LEVELS = ["未测", "初学", "熟练", "精通"]
const TONE = ["bg-foreground/14", "bg-brand/35", "bg-brand/65", "bg-brand"]

/** The ability profile as heat in the right margin: a row per ability, a block
 *  per knowledge cell, nothing written on the page. */
export function AbilityRail() {
  const { abilities, route } = useCourse()
  const [hovered, setHovered] = React.useState<Ability | null>(null)

  const cells = React.useMemo(
    () => (route?.nodes ?? []).flatMap((node) => node.cells),
    [route],
  )

  if (abilities.length === 0) return null

  return (
    <>
      <aside
        aria-label="能力画像"
        className="group/rail fixed top-1/2 right-4 z-20 hidden -translate-y-1/2 flex-col gap-[10px] opacity-90 transition-opacity duration-300 hover:opacity-100 lg:flex"
        onMouseLeave={() => setHovered(null)}
      >
        {abilities.map((ability) => {
          const own = cells.filter((cell) => cell.abilityId === ability.id)
          const shown = own.slice(0, 6)
          return (
            <button
              key={ability.id}
              type="button"
              aria-label={ability.title}
              onMouseEnter={() => setHovered(ability)}
              onFocus={() => setHovered(ability)}
              onBlur={() => setHovered(null)}
              className="flex h-[16px] items-center gap-[4px] rounded-[4px] px-1 transition-colors duration-200 hover:bg-accent/70 focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none"
            >
              {shown.length === 0 ? (
                <span className="h-[1px] w-[9px] rounded-full bg-foreground/20" />
              ) : (
                shown.map((cell) => (
                  <span
                    key={cell.id}
                    className={cn("size-[9px] rounded-[2px]", TONE[Math.min(3, Math.max(0, cell.mastery))])}
                  />
                ))
              )}
              {own.length > 6 ? (
                <span className="ml-[1px] font-mono text-[9px] text-foreground/50">+{own.length - 6}</span>
              ) : null}
            </button>
          )
        })}
      </aside>

      {hovered ? <AbilityCard ability={hovered} cells={cells.filter((cell) => cell.abilityId === hovered.id)} /> : null}
    </>
  )
}

function AbilityCard({ ability, cells }: { ability: Ability; cells: Cell[] }) {
  const average =
    cells.length > 0 ? cells.reduce((sum, cell) => sum + cell.mastery, 0) / cells.length : 0
  const percent = Math.round((average / 3) * 100)
  const level = LEVELS[Math.min(3, Math.max(0, Math.round(average)))]
  const count = (value: number) => cells.filter((cell) => cell.mastery === value).length
  const nodes = new Set(cells.map((cell) => cell.nodeId)).size

  return (
    <div className="pointer-events-none fixed top-1/2 right-[92px] z-20 hidden w-[248px] -translate-y-1/2 lg:block">
      <div className="pane rounded-[14px] px-4 py-3.5">
        <div className="font-mono text-[9px] tracking-[0.1em] text-muted-foreground/70">
          能力 {String(ability.seq).padStart(2, "0")}
        </div>

        <h3 className="mt-1.5 text-[15px] leading-snug text-foreground">{ability.title}</h3>

        {ability.description ? (
          <p className="mt-1.5 text-[11px] leading-[1.65] text-muted-foreground">{ability.description}</p>
        ) : null}

        <div className="mt-3.5 h-[2px] w-full overflow-hidden rounded-full bg-foreground/10">
          <div className="h-full rounded-full bg-brand/80" style={{ width: `${percent}%` }} />
        </div>

        <div className="mt-2 flex items-baseline justify-between">
          <span className="text-[11px] text-muted-foreground">{level}</span>
          <span className="font-mono text-[11px] tabular-nums text-foreground/75">{percent}%</span>
        </div>

        {cells.length > 0 ? (
          <>
            <div className="mt-3 flex items-center gap-3 text-[10px] text-muted-foreground/85">
              <span>初学 {count(1)}</span>
              <span>熟练 {count(2)}</span>
              <span>精通 {count(3)}</span>
            </div>
            <div className="mt-1.5 text-[10px] text-muted-foreground/70">
              覆盖 {nodes} 节 · {cells.length} 个格子
            </div>
          </>
        ) : (
          <p className="mt-3 text-[10px] text-muted-foreground/70">还没有格子归属这项能力。</p>
        )}
      </div>
    </div>
  )
}
