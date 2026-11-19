import { useState } from 'react'

import { useCopy } from '../i18n/i18n'
import { Card, Chip, Section } from './ui'

// CopyButton copies a command and says so for a moment.
export function CopyButton({ text }: { text: string }) {
  const t = useCopy()
  const [copied, setCopied] = useState(false)
  return (
    <button
      type="button"
      onClick={() => {
        void navigator.clipboard.writeText(text).then(() => {
          setCopied(true)
          setTimeout(() => setCopied(false), 1500)
        })
      }}
      className="shrink-0 rounded-full border border-line px-3 py-1 text-xs text-muted hover:border-muted hover:text-text"
    >
      {copied ? t.start.copied : t.start.copy}
    </button>
  )
}

export function Command({ text }: { text: string }) {
  return (
    <div className="flex items-center gap-3 rounded-2xl border border-line bg-ink px-4 py-3 font-mono text-sm">
      <span className="text-muted">$</span>
      <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap">{text}</code>
      <CopyButton text={text} />
    </div>
  )
}

export function QuickStart() {
  const t = useCopy()
  const [tab, setTab] = useState(0)
  const steps = t.start.steps[tab] ?? []
  return (
    <Section id="start">
      <div className="flex flex-wrap items-end justify-between gap-6">
        <div className="max-w-2xl">
          <p className="eyebrow">{t.start.kicker}</p>
          <h2 className="display mt-3 text-4xl md:text-5xl">{t.start.title}</h2>
          <p className="mt-4 text-lg leading-relaxed text-muted">{t.start.sub}</p>
        </div>
        <div role="tablist" aria-label={t.start.kicker} className="flex rounded-full border border-line bg-panel p-1">
          {t.start.tabs.map((label, i) => (
            <button
              key={label}
              type="button"
              role="tab"
              aria-selected={i === tab}
              onClick={() => setTab(i)}
              className={`rounded-full px-4 py-2 text-sm transition ${i === tab ? 'bg-text font-semibold text-ink' : 'text-muted hover:text-text'}`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>
      <div role="tabpanel" className="mt-10 grid gap-5 md:grid-cols-3">
        {steps.map(([title, body, command], i) => (
          <Card key={title} className="p-7">
            <span className="grid size-8 place-items-center rounded-full border border-amber/50 text-sm font-semibold text-amber">{i + 1}</span>
            <h3 className="mt-5 text-lg font-semibold">{title}</h3>
            <p className="mt-2 leading-relaxed text-muted">{body}</p>
            {command && <div className="mt-4"><Command text={command} /></div>}
          </Card>
        ))}
      </div>
      <div className="mt-8 flex flex-wrap items-center gap-2">
        <span className="mr-1 text-sm text-muted">{t.start.backends}</span>
        {t.start.backendList.map((b) => <Chip key={b}>{b}</Chip>)}
      </div>
    </Section>
  )
}
