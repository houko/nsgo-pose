export const LANGS = [
  { code: 'zh', path: '/', html: 'zh-Hans', label: '简体中文' },
  { code: 'en', path: '/en', html: 'en', label: 'English' },
  { code: 'ja', path: '/ja', html: 'ja', label: '日本語' },
] as const

export type Lang = (typeof LANGS)[number]['code']

// langOf is the language a path is in: its first segment, else Chinese, the site's root.
export function langOf(pathname: string): (typeof LANGS)[number] {
  const first = pathname.split('/').filter(Boolean)[0]
  return LANGS.find((l) => l.code === first) ?? LANGS[0]
}
