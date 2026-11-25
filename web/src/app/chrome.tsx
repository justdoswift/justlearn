import * as React from "react"

import { cn } from "@/lib/utils"

/** The ✦ mark: four points, drawn, not a glyph from a font. */
export function Mark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 16 16" aria-hidden="true" className={cn("size-3.5", className)}>
      <path
        d="M8 .7c.46 3.7 3.6 6.84 7.3 7.3-3.7.46-6.84 3.6-7.3 7.3-.46-3.7-3.6-6.84-7.3-7.3C4.4 7.54 7.54 4.4 8 .7Z"
        fill="currentColor"
      />
    </svg>
  )
}

/** A mark that does something: quiet until the pointer is on it. */
export function GlyphButton({
  label,
  onClick,
  className,
  children,
  ...props
}: React.ComponentProps<"button"> & { label: string }) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onClick={onClick}
      className={cn(
        "flex size-8 items-center justify-center rounded-lg text-muted-foreground transition-colors duration-150",
        "hover:bg-accent hover:text-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none",
        className,
      )}
      {...props}
    >
      {children}
    </button>
  )
}

/** The one wordmark on the page, set in the serif it uses for lessons. */
export function Wordmark({ className }: { className?: string }) {
  return (
    <span className={cn("font-serif text-[16px] tracking-[0.06em] text-foreground/90", className)}>
      JustLearn
    </span>
  )
}
