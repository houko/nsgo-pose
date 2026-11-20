import desktopEn from '../../../screenshots/desktop-en.png'
import desktopJa from '../../../screenshots/desktop-ja.png'
import desktopZh from '../../../screenshots/desktop-zh-Hans.png'
import phoneEn from '../../../screenshots/phone-home-en.png'
import phoneJa from '../../../screenshots/phone-home-ja.png'
import phoneZh from '../../../screenshots/phone-home-zh-Hans.png'
import type { Lang } from './i18n/langs'

// screenshots are the app in each of the site's languages, from the repository's screenshots/: its phone home and its desktop studio.
export const screenshots: Record<Lang, { phone: string; desktop: string }> = {
  zh: { phone: phoneZh, desktop: desktopZh },
  en: { phone: phoneEn, desktop: desktopEn },
  ja: { phone: phoneJa, desktop: desktopJa },
}
