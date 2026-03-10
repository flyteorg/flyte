// src/lib/urlUtils.test.ts
import { describe, it, expect } from 'vitest'
import {
  getUpdatedScopedPath,
  getUpdatedDomainPath,
  getUpdatedProjectPath,
  getS3ConsoleUrl,
} from './urlUtils'

describe('urlUtils', () => {
  describe('getUpdatedDomainPath', () => {
    it('replaces the domain segment and keeps list route intact', () => {
      expect(
        getUpdatedDomainPath('/v2/domain/old/project/p1/runs', 'new'),
      ).toBe('/v2/domain/new/project/p1/runs')
    })

    it('replaces the domain segment and trims detail routes back to section list', () => {
      expect(
        getUpdatedDomainPath('/v2/domain/old/project/p1/runs/abc123', 'new'),
      ).toBe('/v2/domain/new/project/p1/runs')
    })

    it('trims even if there are multiple extra segments after the section', () => {
      expect(
        getUpdatedDomainPath(
          '/v2/domain/old/project/p1/runs/abc123/logs',
          'new',
        ),
      ).toBe('/v2/domain/new/project/p1/runs')
    })
  })

  describe('getUpdatedProjectPath', () => {
    it('replaces the project segment and keeps list route intact', () => {
      expect(
        getUpdatedProjectPath('/v2/domain/d1/project/old/tasks', 'newP'),
      ).toBe('/v2/domain/d1/project/newP/tasks')
    })

    it('replaces the project segment and trims detail routes back to section list', () => {
      expect(
        getUpdatedProjectPath('/v2/domain/d1/project/old/tasks/t42', 'newP'),
      ).toBe('/v2/domain/d1/project/newP/tasks')
    })
  })

  describe('getUpdatedScopedPath', () => {
    it('can update both domain and project in one call', () => {
      expect(
        getUpdatedScopedPath('/v2/domain/old/project/p1/runs', {
          domainId: 'newD',
          projectId: 'newP',
        }),
      ).toBe('/v2/domain/newD/project/newP/runs')
    })

    it('updates both and trims detail routes back to section list', () => {
      expect(
        getUpdatedScopedPath('/v2/domain/old/project/p1/runs/r123', {
          domainId: 'newD',
          projectId: 'newP',
        }),
      ).toBe('/v2/domain/newD/project/newP/runs')
    })

    it('leaves unrelated routes alone except for the requested replacement', () => {
      // No /project segment => just domain replacement (via segment lookup or regex fallback)
      expect(
        getUpdatedScopedPath('/v2/domain/old/settings', { domainId: 'new' }),
      ).toBe('/v2/domain/new/settings')
    })

    it('falls back to regex replacement when segments are missing', () => {
      // Project missing but /domain exists -> domain regex fallback should still work
      expect(
        getUpdatedScopedPath('/v2/domain/old/runs/abc', { domainId: 'newD' }),
      ).toBe('/v2/domain/newD/runs/abc')
    })

    it('does not trim when there is no section after /project/:projectId', () => {
      expect(
        getUpdatedScopedPath('/v2/domain/old/project/p1', { domainId: 'newD' }),
      ).toBe('/v2/domain/newD/project/p1')
    })

    it('does not trim when already on section list route', () => {
      expect(
        getUpdatedScopedPath('/v2/domain/old/project/p1/apps', {
          domainId: 'newD',
        }),
      ).toBe('/v2/domain/newD/project/p1/apps')
    })
  })

  describe('getS3ConsoleUrl', () => {
    const globalBase = 'https://s3.console.aws.amazon.com'

    it('returns S3 console object URL for s3:// bucket/key (file)', () => {
      const uri = 's3://test-bucket/folder1/folder two/file-name.txt'
      const url = getS3ConsoleUrl(uri)
      expect(url).toBe(
        `${globalBase}/s3/object/test-bucket?prefix=folder1%2Ffolder%20two%2Ffile-name.txt`,
      )
    })

    it('returns null for non-S3 URI', () => {
      expect(getS3ConsoleUrl('https://example.com/path')).toBeNull()
      expect(getS3ConsoleUrl('gs://bucket/key')).toBeNull()
      expect(getS3ConsoleUrl('')).toBeNull()
      expect(getS3ConsoleUrl(undefined)).toBeNull()
    })

    it('returns object URL for s3:// with single path segment (file)', () => {
      expect(getS3ConsoleUrl('s3://my-bucket/file.txt')).toBe(
        `${globalBase}/s3/object/my-bucket?prefix=file.txt`,
      )
    })

    it('returns bucket browser URL for directory (key ends with /)', () => {
      expect(getS3ConsoleUrl('s3://my-bucket/folder/')).toBe(
        `${globalBase}/s3/buckets/my-bucket?prefix=folder/&showversions=false`,
      )
    })

    it('returns bucket browser URL for directory without trailing slash (adds / and literal slashes in prefix)', () => {
      // Pass isDirectory: true explicitly so the test does not rely on the extensionless-key heuristic
      const uri =
        's3://union-oc-production-demo/se/demo/flytesnacks/development/r96xprgbs4l5nqkjn687/5azce9tr0xyhis1w9qpwiz5a4/1/ek/r96xprgbs4l5nqkjn687-5azce9tr0xyhis1w9qpwiz5a4-0/08eb75015590e918d21d7312bc25990c/tmpj_woph03'
      const url = getS3ConsoleUrl(uri, { isDirectory: true })
      expect(url).toBe(
        `${globalBase}/s3/buckets/union-oc-production-demo?prefix=se/demo/flytesnacks/development/r96xprgbs4l5nqkjn687/5azce9tr0xyhis1w9qpwiz5a4/1/ek/r96xprgbs4l5nqkjn687-5azce9tr0xyhis1w9qpwiz5a4-0/08eb75015590e918d21d7312bc25990c/tmpj_woph03/&showversions=false`,
      )
    })

    it('returns bucket URL for bucket root (no key)', () => {
      expect(getS3ConsoleUrl('s3://my-bucket')).toBe(
        `${globalBase}/s3/buckets/my-bucket?showversions=false`,
      )
      expect(getS3ConsoleUrl('s3://my-bucket/')).toBe(
        `${globalBase}/s3/buckets/my-bucket?showversions=false`,
      )
    })

    it('uses bucket browser when isDirectory option is true (multipart blob)', () => {
      const uri = 's3://my-bucket/path/to/multipart.blob'
      expect(getS3ConsoleUrl(uri, { isDirectory: true })).toBe(
        `${globalBase}/s3/buckets/my-bucket?prefix=path/to/multipart.blob/&showversions=false`,
      )
    })
  })
})
