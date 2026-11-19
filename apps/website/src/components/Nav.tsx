import { Link, useLocation } from '@tanstack/react-router'

import { useCopy } from '../i18n/i18n'
import { LANGS, langOf } from '../i18n/langs'
import { site } from '../site'
import { GitHubIcon, GlobeIcon } from './icons'

export function Nav() {
  const t = useCopy()
  const lang = langOf(useLocation().pathname)
  return (
    <header className="sticky top-0 z-30 border-b border-line bg-ink/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-6xl items-center gap-6 px-6 py-3">
        <a href="#top" className="flex items-baseline gap-1.5">
          <span className="text-xl font-bold">Mirra</span>
          <span className="text-sm text-muted">镜衣</span>
        </a>
        <nav aria-label={t.nav.label} className="hidden gap-5 text-[15px] text-muted md:flex">
          <a href="#how" className="hover:text-text">{t.nav.how}</a>
          <a href="#why" className="hover:text-text">{t.nav.why}</a>
          <a href="#start" className="hover:text-text">{t.nav.start}</a>
          <a href="#faq" className="hover:text-text">{t.nav.faq}</a>
        </nav>
        <div className="ml-auto flex items-center gap-2">
          <details className="relative">
            <summary aria-label={t.nav.language} className="flex cursor-pointer items-center gap-2 rounded-full border border-line px-3.5 py-2 text-sm hover:border-muted">
              <GlobeIcon />
              {lang.label}
            </summary>
            <ul className="absolute right-0 mt-2 w-36 overflow-hidden rounded-2xl border border-line bg-panel py-1 text-sm shadow-xl">
              {LANGS.map((l) => (
                <li key={l.code}>
                  <Link to={l.path} hrefLang={l.html} className={`block px-4 py-2 hover:bg-raised ${l.code === lang.code ? 'text-amber' : ''}`}>
                    {l.label}
                  </Link>
                </li>
              ))}
            </ul>
          </details>
          <a href={site.repo} target="_blank" rel="noopener" className="hidden items-center gap-2 rounded-full border border-line px-3.5 py-2 text-sm hover:border-muted sm:flex">
            <GitHubIcon />
            GitHub
            {site.stars !== undefined && site.stars > 0 && <span className="border-l border-line pl-2 text-amber">★ {site.stars}</span>}
          </a>
          <a href="#download" className="rounded-full bg-amber px-4 py-2 text-sm font-semibold text-amber-ink">
            {t.nav.download}
          </a>
        </div>
      </div>
    </header>
  )
}
