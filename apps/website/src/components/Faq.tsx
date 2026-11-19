import { useCopy } from '../i18n/i18n'
import { links } from '../site'
import { Section } from './ui'

export function Faq() {
  const t = useCopy()
  return (
    <Section id="faq">
      <div className="grid gap-12 md:grid-cols-[1fr_1.8fr]">
        <div>
          <p className="eyebrow">{t.faq.kicker}</p>
          <h2 className="display mt-3 text-4xl md:text-5xl">{t.faq.title}</h2>
          <p className="mt-4 text-muted">
            <a href={links.discussions} target="_blank" rel="noopener" className="hover:text-text">{t.faq.sub}</a>
          </p>
        </div>
        <div className="border-t border-line">
          {t.faq.items.map(([q, a], i) => (
            <details key={q} open={i === 0} className="group border-b border-line">
              <summary className="flex cursor-pointer items-center justify-between gap-6 py-5 text-lg font-medium">
                {q}
                <span aria-hidden="true" className="grid size-7 shrink-0 place-items-center rounded-full border border-line text-amber group-open:rotate-45 transition">+</span>
              </summary>
              <p className="pb-6 leading-relaxed text-muted">{a}</p>
            </details>
          ))}
        </div>
      </div>
    </Section>
  )
}
