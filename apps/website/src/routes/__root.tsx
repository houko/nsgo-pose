import { HeadContent, Outlet, Scripts, createRootRoute, useLocation } from '@tanstack/react-router'

import icon from '../../../mirra/assets/Icon.png'
import { langOf } from '../i18n/langs'
import css from '../styles/app.css?url'

export const Route = createRootRoute({
  head: () => ({
    meta: [{ charSet: 'utf-8' }, { name: 'viewport', content: 'width=device-width, initial-scale=1' }, { name: 'theme-color', content: '#0e0d0c' }],
    links: [
      { rel: 'stylesheet', href: css },
      { rel: 'icon', type: 'image/png', href: icon },
      { rel: 'apple-touch-icon', href: icon },
    ],
  }),
  component: Root,
})

function Root() {
  const lang = langOf(useLocation().pathname)
  return (
    <html lang={lang.html}>
      <head>
        <HeadContent />
      </head>
      <body>
        <Outlet />
        <Scripts />
      </body>
    </html>
  )
}
