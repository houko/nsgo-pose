import { useCopy } from '../i18n/i18n'
import { links, site } from '../site'
import { GlobeIcon, LaptopIcon, PhoneIcon } from './icons'
import { Button, Card, Section } from './ui'

// The end of each platform's file name in a release, after mirra-<version>- (.github/workflows/package.yml); iOS is built from source and the web app is published with this site.
const files = [undefined, 'android-arm64.apk', 'macos-arm64.dmg', 'windows-x64-setup.exe', undefined]

// file is the address of the latest release's file of a platform, when the release holds one: a release may leave a platform out, such as Android before its signing key is set up.
function file(i: number): string | undefined {
  const suffix = files[i]
  if (!suffix || !site.release) return undefined
  return site.release.files[`mirra-${site.release.tag.replace(/^v/, '')}-${suffix}`]
}

export function Download() {
  const t = useCopy()
  return (
    <Section id="download">
      <p className="eyebrow">{t.download.kicker}</p>
      <h2 className="display mt-3 text-4xl md:text-5xl">{t.download.title}</h2>
      <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        {t.download.platforms.map(([name, requirement, action], i) => {
          const Icon = i < 2 ? PhoneIcon : i < 4 ? LaptopIcon : GlobeIcon
          const href = i === 0 ? links.development : i === 4 ? links.webApp : file(i)
          return (
            <Card key={name} className="flex flex-col p-6">
              <Icon className="size-6 text-muted" />
              <h3 className="mt-4 text-xl font-semibold">{name}</h3>
              <p className="mt-1 text-sm text-muted">{requirement}</p>
              {href ? (
                <Button href={href} className="mt-6 w-full">{action}</Button>
              ) : (
                <span aria-disabled="true" className="mt-6 inline-flex w-full items-center justify-center rounded-full border border-line px-6 py-3 text-[15px] text-muted">
                  {t.download.soon}
                </span>
              )}
            </Card>
          )
        })}
      </div>
      {!site.release && (
        <p className="mt-6 text-sm text-muted">
          <a href={links.releases} target="_blank" rel="noopener" className="underline decoration-line underline-offset-4 hover:text-text">{t.download.noRelease}</a>
        </p>
      )}
      <p className="mt-3 text-sm text-muted">{t.download.note}</p>
    </Section>
  )
}
