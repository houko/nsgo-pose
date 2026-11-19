import { copies } from './copy'
import type { Lang } from './langs'

// head is a page's title and description in its language, for its route's head.
export function head(lang: Lang) {
  const { title, description } = copies[lang].meta
  return {
    meta: [
      { title },
      { name: 'description', content: description },
      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { property: 'og:type', content: 'website' },
    ],
  }
}
