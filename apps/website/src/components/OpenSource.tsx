import { useCopy } from '../i18n/i18n'
import type { Lang } from '../i18n/langs'
import { screenshots } from '../screenshots'
import { links, site } from '../site'
import { Command } from './QuickStart'
import { Button, Card, Section } from './ui'

export function OpenSource({ lang }: { lang: Lang }) {
  const t = useCopy()
  const communityLinks = [links.issues, links.discussions, links.development]
  return (
    <Section id="oss">
      <Card className="grid overflow-hidden md:grid-cols-[1.2fr_1fr]">
        <div className="p-10 md:p-12">
          <p className="eyebrow">{t.oss.kicker}</p>
          <h2 className="display mt-3 text-4xl">{t.oss.title}</h2>
          <p className="mt-4 text-lg leading-relaxed text-muted">{t.oss.sub}</p>
          <div className="mt-6"><Command text={`git clone ${site.repo}.git`} /></div>
          <div className="mt-6 flex flex-wrap gap-3">
            <Button href={site.repo} kind="light">{t.oss.star}</Button>
            <Button href={links.development} kind="outline">{t.oss.guide}</Button>
          </div>
        </div>
        <div className="relative hidden items-center overflow-hidden bg-raised md:flex">
          <img src={screenshots[lang].desktop} alt="" width={1440} height={917} loading="lazy" className="ml-8 w-[160%] max-w-none rounded-2xl border border-line shadow-2xl" />
        </div>
      </Card>
      <div className="mt-5 grid gap-5 md:grid-cols-3">
        {t.oss.community.map(([title, body, cta], i) => (
          <Card key={title} className="p-7">
            <h3 className="text-lg font-semibold">{title}</h3>
            <p className="mt-2 leading-relaxed text-muted">{body}</p>
            <a href={communityLinks[i]} target="_blank" rel="noopener" className="mt-4 inline-block text-amber hover:underline">{cta}</a>
          </Card>
        ))}
      </div>
      <div className="mt-5 flex flex-wrap items-center gap-4 rounded-3xl border border-dashed border-line px-7 py-5">
        <span className="font-semibold">{t.oss.thanks}</span>
        {site.contributors.length > 0 ? (
          <div className="flex -space-x-2">
            {site.contributors.map((c) => (
              <a key={c.login} href={c.url} target="_blank" rel="noopener" title={c.login}>
                <img src={`${c.avatar}&s=80`} alt={c.login} width={36} height={36} loading="lazy" className="size-9 rounded-full border-2 border-ink" />
              </a>
            ))}
          </div>
        ) : (
          <a href={links.development} target="_blank" rel="noopener" className="text-amber hover:underline">{t.oss.firstContributor}</a>
        )}
      </div>
    </Section>
  )
}
