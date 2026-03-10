/**
 * Decodes a URI string (e.g. %20 -> space) so the form stores and displays
 * human-readable paths. Returns the original string if decoding fails.
 */
export function decodeBlobUri(uri: string): string {
  if (typeof uri !== 'string' || uri === '') return uri
  try {
    return decodeURIComponent(uri)
  } catch {
    return uri
  }
}
