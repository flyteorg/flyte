'use client'

import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'

/** Placeholder for advanced metrics (licensed edition). Uses shared LicensedEditionPlaceholder with "Advanced metrics" title. */
export function AdvancedMetricsPlaceholder() {
  return <LicensedEditionPlaceholder fullWidth title="Advanced metrics" />
}
