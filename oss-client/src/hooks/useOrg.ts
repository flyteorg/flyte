import { BASE_ADMIN_API } from '@/lib/apiUtils'
import { useEffect, useState } from 'react'
import { isBrowser as getIsBrowser, getWindow } from '@/lib/windowUtils'

export const useOrg = () => {
  const [org, setOrg] = useState<string>('')

  useEffect(() => {
    const isDevBox = BASE_ADMIN_API?.includes('localhost')
    // Server-side: Check Node.js environment variables
    const isBrowser = getIsBrowser()
    if (!isBrowser) {
      console.log('SSR - checking for ORG override')
      const orgOverride = process.env.UNION_ORG_OVERRIDE
      if (orgOverride) {
        setOrg(orgOverride)
        return
      }

      // Server-side: cannot access window.location, return default
      // This should not be called during SSR for API requests
      setOrg('')
      return
    }
    const w = getWindow()

    // Check for runtime environment variable override first
    const unionOrgOverride = w?.__ENV?.UNION_ORG_OVERRIDE
    if (unionOrgOverride) {
      setOrg(unionOrgOverride)
      return
    }

    if (isDevBox) {
      setOrg('testorg')
      return
    }
    if (w?.location.hostname.replace) {
      // Legacy, BYOC environments extract the organization name from the hostname
      setOrg(
        w?.location?.hostname?.replace?.('localhost.', '')?.split('.')?.[0],
      )
    }
  }, [])

  return org
}
