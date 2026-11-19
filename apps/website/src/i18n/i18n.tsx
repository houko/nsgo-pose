import i18next from 'i18next'
import type { ReactNode } from 'react'
import { I18nextProvider, initReactI18next, useTranslation } from 'react-i18next'

import { type Copy, copies } from './copy'
import type { Lang } from './langs'

// One i18next instance per language, initialized synchronously so a page prerenders with its words in place.
const instances = Object.fromEntries(
  (Object.keys(copies) as Lang[]).map((lng) => {
    const i18n = i18next.createInstance()
    void i18n.use(initReactI18next).init({
      lng,
      resources: { [lng]: { site: copies[lng] } },
      defaultNS: 'site',
      initAsync: false,
      interpolation: { escapeValue: false },
      showSupportNotice: false,
    })
    return [lng, i18n]
  }),
) as Record<Lang, typeof i18next>

export function LangProvider({ lang, children }: { lang: Lang; children: ReactNode }) {
  return <I18nextProvider i18n={instances[lang]}>{children}</I18nextProvider>
}

// useCopy is the page's words in its language, as one typed object.
export function useCopy(): Copy {
  const { i18n } = useTranslation()
  return i18n.getResourceBundle(i18n.language, 'site') as Copy
}
