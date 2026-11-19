import { createFileRoute, notFound } from '@tanstack/react-router'

import phoneEn from '../../../../screenshots/phone-home-en.png'
import phoneZh from '../../../../screenshots/phone-home-zh-Hans.png'
import { face } from '../components/ui'
import { copies } from '../i18n/copy'
import { LANGS, type Lang } from '../i18n/langs'

// The 1200x630 card each language's Open Graph image is taken from (scripts/og.sh). It is not prerendered or in the sitemap: only the dev server serves it.
export const Route = createFileRoute('/og/$lang')({
  loader: ({ params }) => {
    if (!LANGS.some((l) => l.code === params.lang)) throw notFound()
  },
  head: () => ({ meta: [{ name: 'robots', content: 'noindex' }] }),
  notFoundComponent: () => <p>Not found</p>,
  component: OgCard,
})

function OgCard() {
  const lang = Route.useParams().lang as Lang
  const t = copies[lang]
  return (
    <div className="relative flex h-[630px] w-[1200px] overflow-hidden bg-ink">
      <div aria-hidden="true" className="absolute inset-0 bg-[radial-gradient(50rem_30rem_at_85%_10%,rgba(227,190,134,0.18),transparent_60%)]" />
      <div className="relative flex w-[700px] flex-col justify-center pl-20">
        <p className="flex items-baseline gap-2">
          <span className="text-3xl font-bold">Mirra</span>
          <span className="text-lg text-muted">镜衣</span>
        </p>
        <p className="eyebrow mt-10">{t.hero.kicker}</p>
        <h1 className="display mt-4 text-[64px]">
          {t.hero.title.map((line, i) => (
            <span key={line} className={`block whitespace-nowrap ${i === t.hero.title.length - 1 ? 'text-amber' : ''}`}>
              {line}
            </span>
          ))}
        </h1>
        <p className="mt-8 text-xl text-muted">{t.footer.tag}</p>
      </div>
      <div className="relative flex-1">
        <img src={lang === 'zh' ? phoneZh : phoneEn} alt="" className="absolute top-14 left-10 w-[300px] rotate-3 rounded-[2rem] border border-line shadow-2xl" />
        <figure className="absolute top-40 -left-6 w-36 -rotate-6 rounded-2xl border border-line bg-panel p-2 shadow-2xl">
          <img src={face(1)} alt="" className="aspect-[4/5] w-full rounded-xl object-cover" />
          <figcaption className="px-1 pt-2 text-[10px] tracking-[0.2em] text-muted">01 FACE</figcaption>
        </figure>
      </div>
    </div>
  )
}
