import { createRouter } from '@tanstack/react-router'

import { routeTree } from './routeTree.gen'

export function getRouter() {
  // Pages serves each page as a folder, so addresses end in a slash, as the canonical links and sitemap give them.
  return createRouter({ routeTree, scrollRestoration: true, trailingSlash: 'always' })
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof getRouter>
  }
}
