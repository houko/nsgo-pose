import type { ReactNode } from 'react'

// Small pieces every section uses.

export function Section({ id, children, className = '' }: { id?: string; children: ReactNode; className?: string }) {
  return (
    <section id={id} className={`border-t border-line ${className}`}>
      <div className="mx-auto max-w-6xl px-6 py-24">{children}</div>
    </section>
  )
}

export function Heading({ kicker, title, sub }: { kicker: string; title: ReactNode; sub?: string }) {
  return (
    <div className="max-w-2xl">
      <p className="eyebrow">{kicker}</p>
      <h2 className="display mt-3 text-4xl md:text-5xl">{title}</h2>
      {sub && <p className="mt-4 text-lg leading-relaxed text-muted">{sub}</p>}
    </div>
  )
}

export function Card({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`rounded-3xl border border-line bg-panel ${className}`}>{children}</div>
}

const buttonStyles = {
  primary: 'bg-amber text-amber-ink shadow-[0_8px_30px_-8px_rgba(227,190,134,0.55)] hover:brightness-105',
  outline: 'border border-line text-text hover:border-muted',
  light: 'bg-text text-ink hover:brightness-95',
}

export function Button({ href, children, kind = 'primary', className = '' }: { href: string; children: ReactNode; kind?: keyof typeof buttonStyles; className?: string }) {
  const external = /^https?:/.test(href)
  return (
    <a
      href={href}
      {...(external ? { target: '_blank', rel: 'noopener' } : {})}
      className={`inline-flex items-center justify-center gap-2 rounded-full px-6 py-3 text-[15px] font-semibold transition ${buttonStyles[kind]} ${className}`}
    >
      {children}
    </a>
  )
}

export function Chip({ children }: { children: ReactNode }) {
  return <span className="rounded-full border border-line px-3 py-1 text-sm text-muted">{children}</span>
}

// face is one of the CC0 model faces the site carries.
export function face(n: number): string {
  return `${import.meta.env.BASE_URL}faces/face-${n}.webp`
}
