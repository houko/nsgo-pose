import { site } from '../site'
import { copies } from './copy'
import { LANGS, type Lang, pageURL } from './langs'

// head is a page's title, description, canonical address, language alternates, Open Graph and Twitter cards and structured data, in its language.
export function head(code: Lang) {
  const lang = LANGS.find((l) => l.code === code) ?? LANGS[0]
  const { title, description } = copies[code].meta
  const url = pageURL(site.url, lang)
  const image = `${site.url}og-${code}.png`
  return {
    meta: [
      { title },
      { name: 'description', content: description },
      { property: 'og:type', content: 'website' },
      { property: 'og:site_name', content: 'Mirra' },
      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { property: 'og:url', content: url },
      { property: 'og:locale', content: lang.og },
      { property: 'og:image', content: image },
      { property: 'og:image:width', content: '1200' },
      { property: 'og:image:height', content: '630' },
      { property: 'og:image:alt', content: title },
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:title', content: title },
      { name: 'twitter:description', content: description },
      { name: 'twitter:image', content: image },
      {
        'script:ld+json': {
          '@context': 'https://schema.org',
          '@type': 'SoftwareApplication',
          name: 'Mirra',
          alternateName: '镜衣',
          description,
          url,
          image,
          applicationCategory: 'MultimediaApplication',
          operatingSystem: 'macOS, iOS, Android, Windows',
          softwareVersion: site.version,
          license: 'https://opensource.org/licenses/MIT',
          isAccessibleForFree: true,
          offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
          sameAs: [site.repo],
          inLanguage: lang.html,
        },
      },
    ],
    links: [
      { rel: 'canonical', href: url },
      ...LANGS.map((l) => ({ rel: 'alternate', hrefLang: l.html, href: pageURL(site.url, l) })),
      { rel: 'alternate', hrefLang: 'x-default', href: site.url },
    ],
  }
}
