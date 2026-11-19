import { useCopy } from '../i18n/i18n'
import { links, site } from '../site'

export function Footer() {
  const t = useCopy()
  const groups: [string, [string, string][]][] = [
    [t.footer.groups[0] ?? '', [[t.footer.links.how, '#how'], [t.footer.links.download, '#download']]],
    [t.footer.groups[1] ?? '', [[t.footer.links.github, site.repo], [t.footer.links.guide, links.development]]],
    [t.footer.groups[2] ?? '', [[t.footer.links.privacy, '#why'], [t.footer.links.sources, links.notice]]],
  ]
  return (
    <footer className="border-t border-line">
      <div className="mx-auto flex max-w-6xl flex-wrap justify-between gap-10 px-6 py-14">
        <div>
          <p className="flex items-baseline gap-1.5"><span className="text-lg font-bold">Mirra</span><span className="text-sm text-muted">镜衣</span></p>
          <p className="mt-3 text-sm text-muted">{t.footer.tag}</p>
          <p className="mt-1 text-sm text-muted">v{site.version} · {t.footer.made}</p>
          <p className="mt-4 text-xs text-muted/70">{t.footer.credit}</p>
        </div>
        <div className="flex gap-14">
          {groups.map(([title, items]) => (
            <div key={title}>
              <p className="text-xs text-muted">{title}</p>
              <ul className="mt-3 space-y-2 text-sm">
                {items.map(([label, href]) => (
                  <li key={label}>
                    <a href={href} {...(/^https?:/.test(href) ? { target: '_blank', rel: 'noopener' } : {})} className="hover:text-amber">{label}</a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </footer>
  )
}
