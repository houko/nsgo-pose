import { useCopy } from '../i18n/i18n'
import { CodeIcon, DevicesIcon, ServerIcon, ShieldIcon } from './icons'
import { Card, Section } from './ui'

const ICONS = [CodeIcon, ServerIcon, ShieldIcon, DevicesIcon]

export function Why() {
  const t = useCopy()
  return (
    <Section id="why">
      <div className="grid gap-12 md:grid-cols-[1fr_1.5fr]">
        <div>
          <p className="eyebrow">{t.why.kicker}</p>
          <h2 className="display mt-3 text-4xl md:text-5xl">
            {t.why.title.map((line) => (
              <span key={line} className="block">{line}</span>
            ))}
          </h2>
          <p className="mt-5 text-lg leading-relaxed text-muted">{t.why.sub}</p>
        </div>
        <div className="grid gap-5 sm:grid-cols-2">
          {t.why.feats.map(([title, body], i) => {
            const Icon = ICONS[i] ?? CodeIcon
            return (
              <Card key={title} className="p-7">
                <Icon className="size-6 text-amber" />
                <h3 className="mt-5 text-lg font-semibold">{title}</h3>
                <p className="mt-2 leading-relaxed text-muted">{body}</p>
              </Card>
            )
          })}
        </div>
      </div>
    </Section>
  )
}
