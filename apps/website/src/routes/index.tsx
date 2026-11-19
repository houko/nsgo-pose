import { createFileRoute } from '@tanstack/react-router'

import { Home } from '../components/Home'
import { head } from '../i18n/head'

export const Route = createFileRoute('/')({ head: () => head('zh'), component: () => <Home lang="zh" /> })
