import { LangProvider } from '../i18n/i18n'
import type { Lang } from '../i18n/langs'
import { Download } from './Download'
import { Faq } from './Faq'
import { Footer } from './Footer'
import { Hero } from './Hero'
import { How } from './How'
import { Nav } from './Nav'
import { OpenSource } from './OpenSource'
import { QuickStart } from './QuickStart'
import { Why } from './Why'

// Home is the landing page in one language. Japanese shows the English screenshots: the app has no Japanese yet.
export function Home({ lang }: { lang: Lang }) {
  const zh = lang === 'zh'
  return (
    <LangProvider lang={lang}>
      <Nav />
      <main>
        <Hero zh={zh} />
        <How />
        <Why />
        <QuickStart />
        <Download />
        <OpenSource zh={zh} />
        <Faq />
      </main>
      <Footer />
    </LangProvider>
  )
}
