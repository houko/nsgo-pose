import { useState } from 'react'

import { useCopy } from '../i18n/i18n'
import type { Lang } from '../i18n/langs'
import { screenshots } from '../screenshots'
import { links, site } from '../site'
import { CheckIcon } from './icons'
import { Button, face } from './ui'

const FACES = [1, 2, 3, 4, 5]

export function Hero({ lang }: { lang: Lang }) {
  const t = useCopy()
  const [current, setCurrent] = useState(1)
  return (
    <section id="top" className="relative overflow-hidden">
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 bg-[radial-gradient(60rem_40rem_at_80%_10%,rgba(227,190,134,0.12),transparent_60%),radial-gradient(40rem_30rem_at_0%_100%,rgba(227,190,134,0.06),transparent_60%)]" />
      <div className="relative mx-auto grid max-w-6xl items-center gap-16 px-6 py-20 md:grid-cols-[1.1fr_1fr] md:py-28">
        <div>
          <a href={links.changelog} target="_blank" rel="noopener" className="inline-flex items-center gap-2 rounded-full border border-line py-1 pr-3 pl-1 text-sm text-muted hover:border-muted">
            <span className="rounded-full bg-amber px-2 py-0.5 text-xs font-semibold text-amber-ink">v{site.version}</span>
            {t.hero.badge} →
          </a>
          <p className="eyebrow mt-8">{t.hero.kicker}</p>
          {/* Each line stays whole, the size following the width, so a longer language's last line does not break. */}
          <h1 className="display mt-4 text-[clamp(2.25rem,4.6vw,4.25rem)]">
            {t.hero.title.map((line, i) => (
              <span key={line} className={`block whitespace-nowrap ${i === t.hero.title.length - 1 ? 'text-amber' : ''}`}>
                {line}
              </span>
            ))}
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-relaxed text-muted">{t.hero.sub}</p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button href="#download">{t.hero.cta}</Button>
            <Button href={site.repo} kind="outline">
              {t.hero.source}
            </Button>
          </div>
          <ul className="mt-6 flex flex-wrap gap-x-5 gap-y-2 text-sm text-muted">
            {t.hero.checks.map((c) => (
              <li key={c} className="flex items-center gap-1.5">
                <CheckIcon className="size-3.5 text-amber" />
                {c}
              </li>
            ))}
          </ul>
        </div>
        <div className="relative mx-auto w-full max-w-sm">
          <img src={screenshots[lang].phone} alt="" width={390} height={844} className="w-full rounded-[2.5rem] border border-line shadow-2xl" />
          <figure className="absolute -left-10 top-16 w-36 -rotate-6 rounded-2xl border border-line bg-panel p-2 shadow-2xl">
            <img src={face(current)} alt="" width={400} height={500} className="aspect-[4/5] w-full rounded-xl object-cover" />
            <figcaption className="px-1 pt-2 text-[10px] tracking-[0.2em] text-muted">01 FACE</figcaption>
          </figure>
          <div className="mt-6 flex items-center justify-center gap-2">
            <span className="mr-1 text-sm text-muted">{t.hero.tryFace} →</span>
            {FACES.map((n) => (
              <button
                key={n}
                type="button"
                aria-label={`${t.hero.faces} ${n}`}
                aria-pressed={n === current}
                onClick={() => setCurrent(n)}
                className={`size-11 overflow-hidden rounded-xl border-2 transition ${n === current ? 'border-amber' : 'border-transparent opacity-70 hover:opacity-100'}`}
              >
                <img src={face(n)} alt="" width={400} height={500} className="size-full object-cover" />
              </button>
            ))}
          </div>
          <p className="mt-2 text-center text-xs text-muted">{t.hero.faces}</p>
        </div>
      </div>
    </section>
  )
}
