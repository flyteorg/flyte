import {
  GPUAccelerator,
  Resources,
  Resources_ResourceName,
  TaskTemplate,
} from '@/gen/flyteidl2/core/tasks_pb'
import type { JsonObject } from '@bufbuild/protobuf'
import stringify from 'safe-stable-stringify'

export type FlattenedCustomItem = {
  name: string
  value: unknown
  copyBtn: boolean
  level?: number
}

/**
 * Flattens task custom (Record<string, unknown>) into a list of section items
 * with level-based nesting up to maxNestLevel. Deeper than maxNestLevel we show
 * the value as formatted JSON. Parent keys (plain objects) get value undefined;
 * leaf values are stringified or formatted as JSON.
 *
 * @param maxNestLevel - Maximum nesting level (0-based). Beyond this, values are shown as JSON. Default 2 (3 levels: 0, 1, 2).
 */
export function flattenCustom(
  obj: Record<string, unknown>,
  maxNestLevel: number = 2,
  level: number = 0,
): FlattenedCustomItem[] {
  const result: FlattenedCustomItem[] = []
  const isPlainObject = (v: unknown) =>
    typeof v === 'object' &&
    v !== null &&
    !Array.isArray(v) &&
    Object.getPrototypeOf(v) === Object.prototype
  for (const [key, value] of Object.entries(obj)) {
    if (isPlainObject(value) && level < maxNestLevel) {
      result.push({ name: key, value: undefined, level, copyBtn: false })
      result.push(
        ...flattenCustom(
          value as Record<string, unknown>,
          maxNestLevel,
          level + 1,
        ),
      )
    } else {
      result.push({
        name: key,
        value:
          typeof value === 'object' && value !== null
            ? stringify(value, null, 2)
            : String(value ?? ''),
        level,
        copyBtn: true,
      })
    }
  }
  return result
}

/** K8s pod_spec.containers[].image; primary container preferred when primaryContainerName is set */
function getImageFromK8sPodSpec(
  podSpec: JsonObject | undefined,
  primaryContainerName: string | undefined,
): string | undefined {
  if (!podSpec || typeof podSpec !== 'object') return undefined
  const containers = podSpec['containers']
  if (!Array.isArray(containers) || containers.length === 0) return undefined
  if (primaryContainerName) {
    const primary = containers.find(
      (c: unknown) =>
        c &&
        typeof c === 'object' &&
        (c as { name?: string }).name === primaryContainerName,
    )
    if (
      primary &&
      typeof primary === 'object' &&
      'image' in primary &&
      typeof (primary as { image: unknown }).image === 'string'
    ) {
      return (primary as { image: string }).image
    }
  }
  const first = containers[0]
  return first &&
    typeof first === 'object' &&
    'image' in first &&
    typeof (first as { image: unknown }).image === 'string'
    ? (first as { image: string }).image
    : undefined
}

/**
 * Resolves the task image for display: container.image for container target,
 * or primary/first container image from podSpec for k8sPod target.
 * Returns undefined for other target types (e.g. sql) or when no image is found.
 */
export function getTaskTargetImage(
  target: TaskTemplate['target'],
): string | undefined {
  if (!target?.case) return undefined
  switch (target.case) {
    case 'container':
      return target.value?.image
    case 'k8sPod':
      return getImageFromK8sPodSpec(
        target.value?.podSpec,
        target.value?.primaryContainerName,
      )
    default:
      return undefined
  }
}

export const getAcceleratorPartitionSizeValue = (
  partitionSizeValue?: GPUAccelerator['partitionSizeValue'],
) => {
  if (!partitionSizeValue) return '-'
  switch (partitionSizeValue?.case) {
    case 'unpartitioned':
      switch (partitionSizeValue.value) {
        case true:
          return 'Unpartitioned'
        case false:
          return 'Partitioned'
      }
    case 'partitionSize':
      return partitionSizeValue.value
    default:
      return '-'
  }
}

/** Result of extracting resource values from a task target (container or k8sPod). */
export type TaskTargetResources = {
  cpuRequests?: string
  cpuLimit?: string
  memoryRequests?: string
  memoryLimit?: string
  ephemeralStorageRequests?: string
  gpuRequests?: string
}

/** Get resource values from a K8s pod spec container (JSON). Uses primary container when name given, else first. */
function getResourcesFromK8sContainer(
  podSpec: JsonObject | undefined,
  primaryContainerName: string | undefined,
): TaskTargetResources {
  if (!podSpec || typeof podSpec !== 'object') return {}
  const containers = podSpec['containers']
  if (!Array.isArray(containers) || containers.length === 0) return {}
  let c: unknown = containers[0]
  if (primaryContainerName) {
    const primary = containers.find(
      (item: unknown) =>
        item &&
        typeof item === 'object' &&
        (item as { name?: string }).name === primaryContainerName,
    )
    if (primary) c = primary
  }
  const container = c as {
    resources?: {
      requests?: Record<string, string>
      limits?: Record<string, string>
    }
  }
  const requests = container?.resources?.requests ?? {}
  const limits = container?.resources?.limits ?? {}
  const getReq = (key: string) =>
    typeof requests[key] === 'string' ? (requests[key] as string) : undefined
  const getLim = (key: string) =>
    typeof limits[key] === 'string' ? (limits[key] as string) : undefined
  const gpuReq =
    getReq('nvidia.com/gpu') ?? getReq('amd.com/gpu') ?? getReq('gpu')
  return {
    cpuRequests: getReq('cpu'),
    cpuLimit: getLim('cpu'),
    memoryRequests: getReq('memory'),
    memoryLimit: getLim('memory'),
    ephemeralStorageRequests: getReq('ephemeral-storage'),
    gpuRequests: gpuReq,
  }
}

/** Get resource values from container target (flyte Resources). */
function getResourcesFromContainerTarget(
  resources: Resources | undefined,
): TaskTargetResources {
  if (!resources) return {}
  const findReq = (name: Resources_ResourceName) =>
    resources.requests?.find((r) => r.name === name)?.value
  const findLim = (name: Resources_ResourceName) =>
    resources.limits?.find((r) => r.name === name)?.value
  return {
    cpuRequests: findReq(Resources_ResourceName.CPU),
    cpuLimit: findLim(Resources_ResourceName.CPU),
    memoryRequests: findReq(Resources_ResourceName.MEMORY),
    memoryLimit: findLim(Resources_ResourceName.MEMORY),
    ephemeralStorageRequests: findReq(Resources_ResourceName.EPHEMERAL_STORAGE),
    gpuRequests: findReq(Resources_ResourceName.GPU),
  }
}

/**
 * Extract resource values from a task target (container or k8sPod).
 * For k8sPod uses primary/first container's resources.requests and resources.limits (K8s keys: cpu, memory, ephemeral-storage, nvidia.com/gpu).
 */
export function getTaskTargetResources(
  target: TaskTemplate['target'],
): TaskTargetResources {
  if (!target?.case) return {}
  switch (target.case) {
    case 'container':
      return getResourcesFromContainerTarget(target.value?.resources)
    case 'k8sPod':
      return getResourcesFromK8sContainer(
        target.value?.podSpec,
        target.value?.primaryContainerName,
      )
    default:
      return {}
  }
}

/**
 * Get value of a flag from container args (e.g. getValueFromArgs(args, '--dest') => value after '--dest').
 * Returns undefined for non-container targets or when not found.
 */
export function getTaskTargetArgValue(
  target: TaskTemplate['target'],
  fieldName: string,
): string | undefined {
  if (target?.case !== 'container') return undefined
  const args = target.value?.args ?? []
  const index = args.indexOf(fieldName)
  return index !== -1 && index + 1 < args.length ? args[index + 1] : undefined
}

/**
 * Normalize a labels/annotations-like value to Record<string, string>.
 * Rejects null, non-objects, and arrays. Only includes entries where both key and value are strings.
 */
function toRecordStringString(
  value: unknown,
): Record<string, string> | undefined {
  if (value == null || typeof value !== 'object' || Array.isArray(value))
    return undefined
  const result: Record<string, string> = {}
  for (const [k, v] of Object.entries(value)) {
    if (typeof k === 'string' && typeof v === 'string') result[k] = v
  }
  return Object.keys(result).length > 0 ? result : undefined
}

/**
 * Get labels from task target. Only k8sPod has metadata.labels; container returns undefined.
 */
export function getTaskTargetLabels(
  target: TaskTemplate['target'],
): Record<string, string> | undefined {
  if (target?.case !== 'k8sPod') return undefined
  return toRecordStringString(target.value?.metadata?.labels)
}

/**
 * Get annotations from task target. Only k8sPod has metadata.annotations; container returns undefined.
 */
export function getTaskTargetAnnotations(
  target: TaskTemplate['target'],
): Record<string, string> | undefined {
  if (target?.case !== 'k8sPod') return undefined
  return toRecordStringString(target.value?.metadata?.annotations)
}
