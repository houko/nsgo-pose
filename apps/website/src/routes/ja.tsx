import { createFileRoute } from '@tanstack/react-router'

import { Home } from '../components/Home'
import { head } from '../i18n/head'

export const Route = createFileRoute('/ja')({ head: () => head('ja'), component: () => <Home lang="ja" /> })
