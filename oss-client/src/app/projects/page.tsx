import { type Metadata } from 'next'

import React from 'react'
import { ProjectsClient } from './ProjectsClient'

export const metadata: Metadata = {
  title: 'Projects',
}

export default function Page() {
  return <ProjectsClient />
}
