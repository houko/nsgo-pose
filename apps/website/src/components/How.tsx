import { useCopy } from '../i18n/i18n'
import { FigureIcon, HangerIcon } from './icons'
import { Card, Chip, Heading, Section, face } from './ui'

export function How() {
  const t = useCopy()
  const visuals = [
    <div key="face" className="grid h-full grid-cols-3 gap-2 p-3">
      {[2, 3, 4].map((n) => (
        <img key={n} src={face(n)} alt="" width={400} height={500} loading="lazy" className="h-full w-full rounded-xl object-cover" />
      ))}
    </div>,
    <div key="outfit" className="flex h-full flex-col items-center justify-center gap-5 p-6 text-amber">
      <HangerIcon className="size-14" />
      <div className="flex flex-wrap justify-center gap-2">{t.how.outfitTags.map((tag) => <Chip key={tag}>{tag}</Chip>)}</div>
    </div>,
    <div key="pose" className="flex h-full flex-col items-center justify-center gap-5 p-6 text-amber">
      <FigureIcon className="size-14" />
      <div className="flex flex-wrap justify-center gap-2">{t.how.angleTags.map((tag) => <Chip key={tag}>{tag}</Chip>)}</div>
    </div>,
  ]
  return (
    <Section id="how">
      <Heading kicker={t.how.kicker} title={t.how.title} />
      <div className="mt-12 grid gap-5 md:grid-cols-3">
        {t.how.steps.map(([title, body], i) => (
          <Card key={title} className="overflow-hidden">
            <div className="h-56 border-b border-line bg-raised">{visuals[i]}</div>
            <div className="p-7">
              <p className="font-serif text-4xl font-bold text-muted/60">0{i + 1}</p>
              <h3 className="display mt-3 text-2xl">{title}</h3>
              <p className="mt-3 leading-relaxed text-muted">{body}</p>
            </div>
          </Card>
        ))}
      </div>
    </Section>
  )
}
