export function getTimeZoneOptions(): string[] {
  if (typeof Intl.supportedValuesOf === 'function') {
    return Intl.supportedValuesOf('timeZone')
  }

  return [
    'UTC',
    'America/Los_Angeles',
    'America/New_York',
    'Europe/London',
    'Europe/Paris',
    'Asia/Tokyo',
  ]
}

export const getBrowserTimezone = () => {
  if (typeof Intl.DateTimeFormat === 'function') {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  }
  return 'UTC'
}
