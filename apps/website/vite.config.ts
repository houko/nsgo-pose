import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import viteReact from '@vitejs/plugin-react'
// defineConfig only types the object; taking it from vite keeps the plugins' types and the config's from one copy of Vite.
import { defineConfig } from 'vite'

import type { SiteData } from './src/site'

const here = path.dirname(fileURLToPath(import.meta.url))

// The site is static: TanStack Start prerenders every page at build time. On GitHub Pages it lives under the repository's path, which CI passes as BASE_PATH; the dev server stays at /.
const base = process.env.BASE_PATH ?? '/'
const repo = process.env.GITHUB_REPOSITORY ?? 'yldm-ai/mirra'

// The app's version, from the packaging configuration the app ships with.
function appVersion(): string {
  const yaml = readFileSync(path.resolve(here, '../mirra/nsgo.yaml'), 'utf8')
  const match = /^version:\s*(\S+)/m.exec(yaml)
  if (!match?.[1]) throw new Error('apps/mirra/nsgo.yaml has no version')
  return match[1]
}

// Stars, contributors and the latest release, read once per build. GitHub's API may be unreachable or rate-limited; the site then shows none of them rather than failing to build.
async function github(): Promise<Pick<SiteData, 'stars' | 'contributors' | 'release'>> {
  const headers: Record<string, string> = { Accept: 'application/vnd.github+json' }
  if (process.env.GITHUB_TOKEN) headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`
  const get = async (url: string) => {
    const res = await fetch(`https://api.github.com/repos/${repo}${url}`, { headers, signal: AbortSignal.timeout(10_000) })
    return res.ok ? res.json() : undefined
  }
  try {
    const [info, contributors, release] = await Promise.all([get(''), get('/contributors?per_page=24'), get('/releases/latest')])
    return {
      stars: typeof info?.stargazers_count === 'number' ? info.stargazers_count : undefined,
      contributors: Array.isArray(contributors)
        ? contributors
            .filter((c: { type?: string }) => c.type === 'User')
            .map((c: { login: string; avatar_url: string; html_url: string }) => ({ login: c.login, avatar: c.avatar_url, url: c.html_url }))
        : [],
      release: release?.html_url ? { tag: String(release.tag_name), url: String(release.html_url) } : undefined,
    }
  } catch {
    return { stars: undefined, contributors: [], release: undefined }
  }
}

const site: SiteData = { version: appVersion(), repo: `https://github.com/${repo}`, ...(await github()) }

export default defineConfig({
  base,
  define: { __SITE__: JSON.stringify(site) },
  server: { fs: { allow: [path.resolve(here, '../..')] } },
  plugins: [
    tailwindcss(),
    tanstackStart({
      router: { basepath: base },
      prerender: { enabled: true, crawlLinks: true, failOnError: true },
    }),
    // viteReact must come after tanstackStart
    viteReact(),
  ],
})
