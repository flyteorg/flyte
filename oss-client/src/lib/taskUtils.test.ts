import { describe, it, expect } from 'vitest'
import { Resources_ResourceName } from '@/gen/flyteidl2/core/tasks_pb'
import type { TaskTemplate } from '@/gen/flyteidl2/core/tasks_pb'
import {
  flattenCustom,
  getTaskTargetImage,
  getAcceleratorPartitionSizeValue,
  getTaskTargetResources,
  getTaskTargetArgValue,
  getTaskTargetLabels,
  getTaskTargetAnnotations,
} from './taskUtils'

/**
 * Tests for taskUtils functions used by TaskDetailsTaskTab and RunDetailsTaskTab:
 * - flattenCustom (task custom config)
 * - getAcceleratorPartitionSizeValue (GPU partition display)
 * - getTaskTargetAnnotations
 * - getTaskTargetArgValue
 * - getTaskTargetImage
 * - getTaskTargetLabels
 * - getTaskTargetResources
 */
describe('taskUtils', () => {
  describe('flattenCustom', () => {
    it('flattens shallow object to list of items with string values', () => {
      const result = flattenCustom({ a: '1', b: '2' })
      expect(result).toHaveLength(2)
      expect(result.map((r) => r.name).sort()).toEqual(['a', 'b'])
      expect(result.every((r) => r.copyBtn === true)).toBe(true)
      expect(result.find((r) => r.name === 'a')?.value).toBe('1')
    })

    it('nests plain objects up to maxNestLevel with level and undefined parent value', () => {
      const result = flattenCustom({ outer: { inner: 'v' } }, 2, 0)
      expect(result.find((r) => r.name === 'outer')?.value).toBeUndefined()
      expect(result.find((r) => r.name === 'outer')?.level).toBe(0)
      expect(result.find((r) => r.name === 'inner')?.value).toBe('v')
      expect(result.find((r) => r.name === 'inner')?.level).toBe(1)
    })

    it('stringifies nested objects beyond maxNestLevel as JSON', () => {
      const result = flattenCustom({ a: { b: { c: 'deep' } } }, 1, 0)
      const deepItem = result.find((r) => r.name === 'b')
      expect(deepItem?.value).toContain('"c"')
      expect(deepItem?.value).toContain('deep')
    })

    it('handles empty object', () => {
      expect(flattenCustom({})).toEqual([])
    })
  })

  describe('getTaskTargetImage', () => {
    it('returns image for container target', () => {
      const target: TaskTemplate['target'] = {
        case: 'container',
        value: { image: 'my-image:tag' },
      }
      expect(getTaskTargetImage(target)).toBe('my-image:tag')
    })

    it('returns undefined for container with no image', () => {
      const target: TaskTemplate['target'] = {
        case: 'container',
        value: {},
      }
      expect(getTaskTargetImage(target)).toBeUndefined()
    })

    it('returns first container image from k8sPod podSpec when no primaryContainerName', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          podSpec: {
            containers: [
              { name: 'c1', image: 'first:1' },
              { image: 'second:2' },
            ],
          },
        },
      }
      expect(getTaskTargetImage(target)).toBe('first:1')
    })

    it('returns primary container image when primaryContainerName is set', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          primaryContainerName: 'main',
          podSpec: {
            containers: [
              { name: 'sidecar', image: 'sidecar:1' },
              { name: 'main', image: 'main:1' },
            ],
          },
        },
      }
      expect(getTaskTargetImage(target)).toBe('main:1')
    })

    it('returns undefined for missing or invalid target', () => {
      expect(getTaskTargetImage(undefined)).toBeUndefined()
      expect(getTaskTargetImage({ case: 'sql', value: {} })).toBeUndefined()
    })
  })

  describe('getAcceleratorPartitionSizeValue', () => {
    it('returns "Unpartitioned" for unpartitioned true', () => {
      expect(
        getAcceleratorPartitionSizeValue({
          case: 'unpartitioned',
          value: true,
        }),
      ).toBe('Unpartitioned')
    })

    it('returns "Partitioned" for unpartitioned false', () => {
      expect(
        getAcceleratorPartitionSizeValue({
          case: 'unpartitioned',
          value: false,
        }),
      ).toBe('Partitioned')
    })

    it('returns partition size string when present', () => {
      expect(
        getAcceleratorPartitionSizeValue({
          case: 'partitionSize',
          value: '2',
        }),
      ).toBe('2')
    })

    it('returns "-" for missing or unknown variant', () => {
      expect(getAcceleratorPartitionSizeValue(undefined)).toBe('-')
      expect(
        getAcceleratorPartitionSizeValue({
          case: 'deviceClass',
          value: 'nvidia',
        } as Parameters<typeof getAcceleratorPartitionSizeValue>[0]),
      ).toBe('-')
    })
  })

  describe('getTaskTargetResources', () => {
    it('extracts resources from container target with Resources (enum name)', () => {
      const target: TaskTemplate['target'] = {
        case: 'container',
        value: {
          resources: {
            requests: [
              { name: Resources_ResourceName.CPU, value: '1000m' },
              { name: Resources_ResourceName.MEMORY, value: '1Gi' },
              { name: Resources_ResourceName.GPU, value: '1' },
              {
                name: Resources_ResourceName.EPHEMERAL_STORAGE,
                value: '2Gi',
              },
            ],
            limits: [
              { name: Resources_ResourceName.CPU, value: '2000m' },
              { name: Resources_ResourceName.MEMORY, value: '2Gi' },
            ],
          },
        },
      }
      const r = getTaskTargetResources(target)
      expect(r.cpuRequests).toBe('1000m')
      expect(r.cpuLimit).toBe('2000m')
      expect(r.memoryRequests).toBe('1Gi')
      expect(r.memoryLimit).toBe('2Gi')
      expect(r.gpuRequests).toBe('1')
      expect(r.ephemeralStorageRequests).toBe('2Gi')
    })

    it('returns empty object for container with no resources', () => {
      const target: TaskTemplate['target'] = { case: 'container', value: {} }
      expect(getTaskTargetResources(target)).toEqual({})
    })

    it('extracts resources from k8sPod podSpec (first container)', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          podSpec: {
            containers: [
              {
                resources: {
                  requests: {
                    cpu: '500m',
                    memory: '512Mi',
                    'ephemeral-storage': '1Gi',
                    'nvidia.com/gpu': '1',
                  },
                  limits: { cpu: '1', memory: '1Gi' },
                },
              },
            ],
          },
        },
      }
      const r = getTaskTargetResources(target)
      expect(r.cpuRequests).toBe('500m')
      expect(r.memoryRequests).toBe('512Mi')
      expect(r.ephemeralStorageRequests).toBe('1Gi')
      expect(r.gpuRequests).toBe('1')
      expect(r.cpuLimit).toBe('1')
      expect(r.memoryLimit).toBe('1Gi')
    })

    it('uses primary container when primaryContainerName is set on k8sPod', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          primaryContainerName: 'main',
          podSpec: {
            containers: [
              { name: 'sidecar', resources: { requests: { cpu: '100m' } } },
              { name: 'main', resources: { requests: { cpu: '2000m' } } },
            ],
          },
        },
      }
      const r = getTaskTargetResources(target)
      expect(r.cpuRequests).toBe('2000m')
    })

    it('returns empty object for non-container/k8sPod target or missing target', () => {
      expect(getTaskTargetResources(undefined)).toEqual({})
      expect(getTaskTargetResources({ case: 'sql', value: {} })).toEqual({})
    })
  })

  describe('getTaskTargetArgValue', () => {
    it('returns value after flag in container args', () => {
      const target: TaskTemplate['target'] = {
        case: 'container',
        value: { args: ['--dest', '/output', '--other'] },
      }
      expect(getTaskTargetArgValue(target, '--dest')).toBe('/output')
      expect(getTaskTargetArgValue(target, '--other')).toBeUndefined()
    })

    it('returns undefined for k8sPod target (args are container-only)', () => {
      const target: TaskTemplate['target'] = { case: 'k8sPod', value: {} }
      expect(getTaskTargetArgValue(target, '--dest')).toBeUndefined()
    })

    it('returns undefined when flag not found or no value after', () => {
      const target: TaskTemplate['target'] = {
        case: 'container',
        value: { args: ['--dest'] },
      }
      expect(getTaskTargetArgValue(target, '--dest')).toBeUndefined()
      expect(getTaskTargetArgValue(target, '--missing')).toBeUndefined()
    })
  })

  describe('getTaskTargetLabels', () => {
    it('returns labels for k8sPod with valid string labels', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            labels: { app: 'my-app', env: 'prod' },
          },
        },
      }
      expect(getTaskTargetLabels(target)).toEqual({
        app: 'my-app',
        env: 'prod',
      })
    })

    it('returns undefined for container target (labels are k8sPod-only)', () => {
      const target: TaskTemplate['target'] = { case: 'container', value: {} }
      expect(getTaskTargetLabels(target)).toBeUndefined()
    })

    it('returns undefined when labels is null or undefined', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: { metadata: {} },
      }
      expect(getTaskTargetLabels(target)).toBeUndefined()
    })

    it('rejects arrays (does not treat as object)', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            labels: [['key', 'val']] as unknown as Record<string, string>,
          },
        },
      }
      expect(getTaskTargetLabels(target)).toBeUndefined()
    })

    it('filters out non-string values and keeps only string entries', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            labels: {
              good: 'yes',
              num: 123 as unknown as string,
              obj: { x: 1 } as unknown as string,
              another: 'ok',
            },
          },
        },
      }
      expect(getTaskTargetLabels(target)).toEqual({
        good: 'yes',
        another: 'ok',
      })
    })

    it('returns undefined when no string entries remain after filtering', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            labels: { a: 1, b: null } as unknown as Record<string, string>,
          },
        },
      }
      expect(getTaskTargetLabels(target)).toBeUndefined()
    })
  })

  describe('getTaskTargetAnnotations', () => {
    it('returns annotations for k8sPod with valid string annotations', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            annotations: { key1: 'v1', key2: 'v2' },
          },
        },
      }
      expect(getTaskTargetAnnotations(target)).toEqual({
        key1: 'v1',
        key2: 'v2',
      })
    })

    it('returns undefined for container target (annotations are k8sPod-only)', () => {
      const target: TaskTemplate['target'] = { case: 'container', value: {} }
      expect(getTaskTargetAnnotations(target)).toBeUndefined()
    })

    it('returns undefined when annotations is null or undefined', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: { metadata: {} },
      }
      expect(getTaskTargetAnnotations(target)).toBeUndefined()
    })

    it('rejects arrays (does not treat as object)', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            annotations: [['k', 'v']] as unknown as Record<string, string>,
          },
        },
      }
      expect(getTaskTargetAnnotations(target)).toBeUndefined()
    })

    it('filters out non-string values and keeps only string entries', () => {
      const target: TaskTemplate['target'] = {
        case: 'k8sPod',
        value: {
          metadata: {
            annotations: {
              desc: 'text',
              count: 42 as unknown as string,
              nested: {} as unknown as string,
            },
          },
        },
      }
      expect(getTaskTargetAnnotations(target)).toEqual({ desc: 'text' })
    })
  })
})
