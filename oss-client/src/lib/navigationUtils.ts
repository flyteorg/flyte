/**
 * Constructs a profile page URL with optional returnTo parameter.
 * If the current pathname is already a settings page, the returnTo parameter is omitted.
 * Otherwise, the current pathname is encoded and added as a returnTo query parameter.
 *
 * @param pathname - The current pathname (e.g., from usePathname())
 * @returns The profile page URL with returnTo parameter if applicable
 */
export function getProfileHref(pathname: string | null | undefined): string {
  // If we're already on a settings page, don't add returnTo
  if (pathname?.startsWith('/settings')) {
    return '/settings/profile'
  }
  // Capture the current pathname as returnTo parameter
  const returnTo = pathname ? encodeURIComponent(pathname) : ''
  return returnTo
    ? `/settings/profile?returnTo=${returnTo}`
    : '/settings/profile'
}
