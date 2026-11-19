export const LANGS = [
  { code: 'zh', path: '/', html: 'zh-Hans', og: 'zh_CN', label: '简体中文' },
  { code: 'en', path: '/en/', html: 'en', og: 'en_US', label: 'English' },
  { code: 'ja', path: '/ja/', html: 'ja', og: 'ja_JP', label: '日本語' },
] as const

export type Lang = (typeof LANGS)[number]['code']

// langOf is the language a path is in: its first segment, else Chinese, the site's root.
export function langOf(pathname: string): (typeof LANGS)[number] {
  const first = pathname.split('/').filter(Boolean)[0]
  return LANGS.find((l) => l.code === first) ?? LANGS[0]
}

// pageURL is a language's page as an absolute address; its path ends in a slash, as Pages serves it.
export function pageURL(siteURL: string, lang: (typeof LANGS)[number]): string {
  return `${siteURL}${lang.path.slice(1)}`
}
