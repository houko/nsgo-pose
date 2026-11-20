import promo from '../../../../screenshots/promo.mp4'
import { useCopy } from '../i18n/i18n'
import { Heading, Section } from './ui'

// Demo is the promo video the README also shows, from the repository's screenshots/; it loads only when played.
export function Demo() {
  const t = useCopy()
  return (
    <Section id="demo">
      <Heading kicker={t.demo.kicker} title={t.demo.title} />
      <video
        src={promo}
        poster={`${import.meta.env.BASE_URL}promo-poster.webp`}
        controls
        muted
        playsInline
        preload="none"
        width={1920}
        height={1080}
        aria-label={t.demo.label}
        className="mt-12 aspect-video w-full rounded-3xl border border-line bg-panel shadow-2xl"
      />
    </Section>
  )
}
