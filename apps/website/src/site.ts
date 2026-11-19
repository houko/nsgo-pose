// What the build knows about the project: the app's version, the site's address and, when GitHub answered, its stars, contributors and latest release.
export type SiteData = {
  version: string
  repo: string
  // url is the site's absolute address with its base path, ending in a slash.
  url: string
  stars?: number
  contributors: { login: string; avatar: string; url: string }[]
  release?: { tag: string; url: string }
}

declare const __SITE__: SiteData

export const site: SiteData = __SITE__

export const links = {
  changelog: `${__SITE__.repo}/blob/main/CHANGELOG.md`,
  development: `${__SITE__.repo}#development`,
  architecture: `${__SITE__.repo}/blob/main/ARCHITECTURE.md`,
  notice: `${__SITE__.repo}/blob/main/NOTICE.md`,
  issues: `${__SITE__.repo}/issues`,
  discussions: `${__SITE__.repo}/discussions`,
  releases: `${__SITE__.repo}/releases`,
}
