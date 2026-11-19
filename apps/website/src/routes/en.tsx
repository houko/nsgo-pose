import { createFileRoute } from '@tanstack/react-router'

import { Home } from '../components/Home'
import { head } from '../i18n/head'

export const Route = createFileRoute('/en')({ head: () => head('en'), component: () => <Home lang="en" /> })
