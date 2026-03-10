type PathScopeUpdates = {
  /** If provided, replaces the `:domain` segment in `/domain/:domain/...` */
  domainId?: string
  /** If provided, replaces the `:projectId` segment in `/project/:projectId/...` */
  projectId?: string
}

/**
 * Returns a pathname updated with the provided scope changes (domain and/or project).
 *
 * Purpose:
 * - When the user switches **domain** and/or **project**, keep them on the same
 *   list-level section when possible.
 * - If the current route appears to be a **detail page** within a section
 *   (i.e. it has extra segments after the section), trim back to the section list.
 *
 * Examples:
 * - `/.../domain/old/project/p1/runs` + { domainId: 'new' } -> `/.../domain/new/project/p1/runs`
 * - `/.../domain/old/project/p1/runs/abc123` + { domainId: 'new' } -> `/.../domain/new/project/p1/runs`
 * - `/.../domain/d1/project/old/tasks/t42` + { projectId: 'newP' } -> `/.../domain/d1/project/newP/tasks`
 */
export const getUpdatedScopedPath = (
  pathName: string,
  updates: PathScopeUpdates,
) => {
  // `usePathname()` returns a pathname without the origin and without query string.
  const segments = pathName.split('/').filter(Boolean)

  // Update /domain/:domainId
  if (updates.domainId) {
    const domainIdx = segments.indexOf('domain')
    if (domainIdx !== -1 && segments[domainIdx + 1]) {
      segments[domainIdx + 1] = updates.domainId
    } else {
      // If we couldn't find a domain segment, fall back to regex replacement.
      // (Keeps behavior stable for unexpected route shapes.)
      return pathName.replace(/\/domain\/[^/]+/, `/domain/${updates.domainId}`)
    }
  }

  // Update /project/:projectId
  if (updates.projectId) {
    const projectIdx = segments.indexOf('project')
    if (projectIdx !== -1 && segments[projectIdx + 1]) {
      segments[projectIdx + 1] = updates.projectId
    } else {
      // If we couldn't find a project segment, fall back to regex replacement.
      return pathName.replace(
        /\/project\/[^/]+/,
        `/project/${updates.projectId}`,
      )
    }
  }

  // If we're on a project list/detail route, trim any detail segments back to the list.
  // Expected shapes:
  // - /.../domain/:domain/project/:projectId/:section
  // - /.../domain/:domain/project/:projectId/:section/:resourceId(/...)
  const projectIdx = segments.indexOf('project')
  const sectionIdx = projectIdx === -1 ? -1 : projectIdx + 2 // after `project` + `:projectId`

  if (sectionIdx !== -1 && segments[sectionIdx]) {
    // If there are segments after the section, we're on a detail page; trim to list page.
    if (segments.length > sectionIdx + 1) {
      segments.splice(sectionIdx + 1)
    }
  }

  return `/${segments.join('/')}`
}

/** Convenience wrapper for domain changes (backwards-compatible with prior usage). */
export const getUpdatedDomainPath = (pathName: string, newDomain: string) =>
  getUpdatedScopedPath(pathName, { domainId: newDomain })

/** Convenience wrapper for project changes. */
export const getUpdatedProjectPath = (pathName: string, newProjectId: string) =>
  getUpdatedScopedPath(pathName, { projectId: newProjectId })

/**
 * Returns the AWS S3 Console URL for an s3:// URI, or null if the string is not a valid s3 URI.
 *
 * @param uri - The s3:// URI (e.g. from a blob's uri field).
 * @param options.isDirectory - When true, the URI is a multipart blob (directory/prefix); open the
 *   bucket browser at that prefix with a trailing slash. When false or omitted, uses heuristics.
 *
 * Directory format: prefix with literal slashes and trailing slash, showversions=false.
 *
 * @see https://docs.aws.amazon.com/AmazonS3/latest/userguide/view-object-overview.html
 * @see https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-storagebrowser.html
 */
export function getS3ConsoleUrl(
  uri: string | undefined,
  options?: { isDirectory?: boolean },
): string | null {
  if (!uri || typeof uri !== 'string') return null
  const trimmed = uri.trim()
  const match = trimmed.match(/^s3:\/\/([^/]+)(?:\/(.*))?$/)
  if (!match) return null
  const [, bucket, key] = match
  if (!bucket) return null
  const keyPart = key ?? ''
  const base = 'https://s3.console.aws.amazon.com'
  const encodedBucket = encodeURIComponent(bucket)
  const isDirectory =
    options?.isDirectory === true ||
    keyPart === '' ||
    keyPart.endsWith('/')

  if (isDirectory) {
    if (keyPart === '') {
      return `${base}/s3/buckets/${encodedBucket}?showversions=false`
    }
    const prefixForDir = keyPart.endsWith('/') ? keyPart : `${keyPart}/`
    const prefixEncoded = encodeURIComponent(prefixForDir).replace(/%2F/g, '/')
    return `${base}/s3/buckets/${encodedBucket}?prefix=${prefixEncoded}&showversions=false`
  }
  const encodedKey = encodeURIComponent(keyPart)
  return `${base}/s3/object/${encodedBucket}?prefix=${encodedKey}`
}
