// Line icons drawn inline, after Lucide's shapes, so the site needs no icon library.

type IconProps = { className?: string }

const base = { fill: 'none', stroke: 'currentColor', strokeWidth: 1.8, strokeLinecap: 'round', strokeLinejoin: 'round' } as const

export const GitHubIcon = ({ className = 'size-4' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
    <path d="M9 18c-4.51 2-5-2-7-2" />
  </svg>
)
export const GlobeIcon = ({ className = 'size-4' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <circle cx="12" cy="12" r="10" />
    <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
    <path d="M2 12h20" />
  </svg>
)
export const CodeIcon = ({ className = 'size-5' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <path d="m16 18 6-6-6-6" />
    <path d="m8 6-6 6 6 6" />
  </svg>
)
export const ServerIcon = ({ className = 'size-5' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <rect width="20" height="8" x="2" y="2" rx="2" />
    <rect width="20" height="8" x="2" y="14" rx="2" />
    <path d="M6 6h.01M6 18h.01" />
  </svg>
)
export const ShieldIcon = ({ className = 'size-5' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" />
    <path d="m9 12 2 2 4-4" />
  </svg>
)
export const DevicesIcon = ({ className = 'size-5' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <rect width="12" height="9" x="2" y="3" rx="1.5" />
    <path d="M5 16h6" />
    <rect width="7" height="11" x="15" y="9" rx="1.5" />
  </svg>
)
export const PhoneIcon = ({ className = 'size-6' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <rect width="12" height="20" x="6" y="2" rx="2" />
    <path d="M12 18h.01" />
  </svg>
)
export const LaptopIcon = ({ className = 'size-6' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <path d="M20 16V7a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v9m16 0H4m16 0 1.28 2.55a1 1 0 0 1-.9 1.45H3.62a1 1 0 0 1-.9-1.45L4 16" />
  </svg>
)
export const HangerIcon = ({ className = 'size-10' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <path d="M12 7a2 2 0 1 0-2-2" />
    <path d="M12 7v1.5L3.2 15a1.5 1.5 0 0 0 .9 2.7h15.8a1.5 1.5 0 0 0 .9-2.7L12 8.5" />
  </svg>
)
export const FigureIcon = ({ className = 'size-10' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <circle cx="12" cy="4.5" r="2" />
    <path d="M12 7v7m0 0-3 7m3-7 3 7M6 9.5l6-1.5 6 1.5" />
  </svg>
)
export const CheckIcon = ({ className = 'size-4' }: IconProps) => (
  <svg viewBox="0 0 24 24" className={className} aria-hidden="true" {...base}>
    <path d="M20 6 9 17l-5-5" />
  </svg>
)
