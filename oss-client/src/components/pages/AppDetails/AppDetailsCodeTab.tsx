'use client'

import { CodeTabContent } from '@/components/CodeTab/CodeTabContent'
import { getAppIdentifier, useAppDetails } from '@/hooks/useApps'
import { useOrg } from '@/hooks/useOrg'
import { useParams } from 'next/navigation'
import React, { useMemo } from 'react'
import { AppDetailsParams } from './types'

export const AppDetailsCodeTab: React.FC = () => {
  const org = useOrg()
  const params = useParams<AppDetailsParams>()

  const { data } = useAppDetails({
    domain: params.domain,
    name: params.appId,
    org,
    projectId: params.project,
  })

  // Extract container from app spec if it's a container payload
  const container = useMemo(() => {
    const appPayload = data?.app?.spec?.appPayload
    if (appPayload?.case === 'container') {
      return appPayload.value
    }
    return undefined
  }, [data?.app?.spec?.appPayload])

  // Use the same logic as useAppDetails
  const appId = useMemo(() => {
    return getAppIdentifier({
      domain: params.domain,
      name: params.appId,
      org,
      project: params.project,
    })
  }, [params.domain, params.appId, params.project, org])

  return (
    <CodeTabContent
      container={container}
      target={appId ? { type: 'appId', value: appId } : undefined}
      noPadding={true}
    />
  )
}
