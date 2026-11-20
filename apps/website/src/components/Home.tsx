import { LangProvider } from '../i18n/i18n'
import type { Lang } from '../i18n/langs'
import { Demo } from './Demo'
import { Download } from './Download'
import { Faq } from './Faq'
import { Footer } from './Footer'
import { Hero } from './Hero'
import { How } from './How'
import { Nav } from './Nav'
import { OpenSource } from './OpenSource'
import { QuickStart } from './QuickStart'
import { Why } from './Why'

// Home is the landing page in one language, with the app's screenshots in it.
export function Home({ lang }: { lang: Lang }) {
  return (
    <LangProvider lang={lang}>
      <Nav />
      <main>
        <Hero lang={lang} />
        <Demo />
        <How />
        <Why />
        <QuickStart />
        <Download />
        <OpenSource lang={lang} />
        <Faq />
      </main>
      <Footer />
    </LangProvider>
  )
}
