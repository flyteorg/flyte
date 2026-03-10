import type { FormState } from '@rjsf/core'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import customParseFormat from 'dayjs/plugin/customParseFormat'
import type { JSONSchema7 } from 'json-schema'
import {
  formatDuration,
  formatIso8601Duration,
  getMillisecondsFromDurationString,
} from './CustomFields/StringField/durationUtils'

import cloneDeep from 'lodash/cloneDeep'
import { isNumber, isObject, isString } from 'lodash'
import duration from 'dayjs/plugin/duration'

dayjs.extend(utc)
dayjs.extend(customParseFormat)
dayjs.extend(duration)

// ---------------------------------------------------------------------------
// Backend / RJSF value normalization (NONETYPE, enums, defaults)
// ---------------------------------------------------------------------------

/** Backend can send Python None as { type: "null" }; normalize to actual null for RJSF. */
function isNoneTypeObject(value: unknown): value is Record<string, unknown> {
  return (
    isObject(value) &&
    !Array.isArray(value) &&
    Object.keys(value).length === 1 &&
    (value as Record<string, unknown>).type === 'null'
  )
}

/**
 * Recursively replace any { type: "null" } (NONETYPE from backend) with null
 * so RJSF and JSON view show null instead of [object Object].
 */
export function normalizeNoneTypeToNull<T>(data: T): T {
  if (isNoneTypeObject(data)) {
    return null as T
  }
  if (Array.isArray(data)) {
    return data.map((item) => normalizeNoneTypeToNull(item)) as T
  }
  if (data != null && isObject(data)) {
    return Object.fromEntries(
      Object.entries(data).map(([k, v]) => [k, normalizeNoneTypeToNull(v)]),
    ) as T
  }
  return data
}

function normalizeDefaultValue(schema: JSONSchema7) {
  if (schema.type === 'string' && schema.format === 'duration') {
    return formatIso8601Duration(`${schema.default}`)
  }
  if (schema.type === 'array') {
    return ensureArrayIfNumericKeys(schema.default) as JSONSchema7['default']
  }
  const raw = schema.default
  return isNoneTypeObject(raw) ? null : raw
}

/** Build initial form data from schema: defaults, first enum value when no default, NONETYPE → null. */
export function getFormDataFromSchemaDefaults(
  schema: JSONSchema7,
): JSONSchema7['default'] {
  if (!schema?.properties) {
    if (schema.default !== undefined) {
      return normalizeDefaultValue(schema)
    }
    // Enum with no default: use first value so form data matches what the dropdown shows
    if (Array.isArray(schema.enum) && schema.enum.length > 0) {
      return schema.enum[0]
    }
    switch (schema.type) {
      case 'object':
        return {}
      case 'array':
        return []
      case 'boolean':
        return false
      default:
        return undefined
    }
  }

  const result = Object.entries(schema.properties).reduce(
    (props, [key, val]) => {
      if (val === true || val === false) {
        return props
      }
      if (val.default !== undefined) {
        const normalized = cloneDeep(normalizeDefaultValue(val))
        return {
          ...props,
          [key]: isNoneTypeObject(normalized) ? null : normalized,
        }
      }
      // For boolean types without a default, default to false
      if (val.type === 'boolean') {
        return { ...props, [key]: false }
      }
      // For enum without default, use first enum value so stored form data matches dropdown
      const entrySchema = val as JSONSchema7
      if (Array.isArray(entrySchema.enum) && entrySchema.enum.length > 0) {
        return { ...props, [key]: entrySchema.enum[0] }
      }
      const nested = getFormDataFromSchemaDefaults(val)
      return { ...props, [key]: isNoneTypeObject(nested) ? null : nested }
    },
    {},
  )
  return normalizeNoneTypeToNull(result)
}

/** If value is a plain object with only numeric keys (0,1,2,...), return as array; else return unchanged.
 * RJSF has no built-in for this; schema defaults and some widgets can produce this shape for array fields. */
function ensureArrayIfNumericKeys(value: unknown): unknown {
  if (value == null || Array.isArray(value)) return value
  if (!isObject(value)) return value
  const keys = Object.keys(value)
  if (keys.length === 0) return value
  const numericKeys = keys.filter((k) => /^\d+$/.test(k))
  if (numericKeys.length !== keys.length) return value
  const indices = numericKeys.map(Number).sort((a, b) => a - b)
  if (indices.some((n, i) => n !== i)) return value
  return indices.map((i) => (value as Record<string, unknown>)[String(i)])
}

// ---------------------------------------------------------------------------
// Schema resolution ($ref)
// ---------------------------------------------------------------------------

/** Resolve #/$defs/Name to the actual definition; return schema unchanged if not a ref or def is not an object. */
function resolveRef(schema: JSONSchema7, root: JSONSchema7): JSONSchema7 {
  const ref = (schema as { $ref?: string }).$ref
  if (!ref || typeof ref !== 'string' || !ref.startsWith('#/$defs/')) {
    return schema
  }
  const name = ref.split('/').pop()
  const defs = root.$defs as Record<string, JSONSchema7 | boolean> | undefined
  if (!name || !defs || !(name in defs)) {
    return schema
  }
  const def = defs[name]
  if (!isObject(def)) {
    // In JSON Schema 7, $defs entries can be boolean; treat non-object defs as unresolvable.
    return schema
  }
  return def as JSONSchema7
}

/** Schema to use for this node (resolved if it was a $ref). */
function getSchema(schema: JSONSchema7, root: JSONSchema7): JSONSchema7 {
  return resolveRef(schema, root)
}

// ---------------------------------------------------------------------------
// convertRjsfInternalFormats
// Normalizes form values so they match what the UI shows and the backend expects:
// - Duration strings → human-readable (e.g. "2h30m")
// - null/undefined at enum fields → first enum value (avoids [..., null] in arrays)
// - Resolves $ref so enums inside $defs are handled
// ---------------------------------------------------------------------------

function getOneOfBranch(
  schema: JSONSchema7,
  formValues: FormState['formData'],
): JSONSchema7 | null {
  const oneOf = (schema as { oneOf?: JSONSchema7[] }).oneOf
  if (!Array.isArray(oneOf) || oneOf.length === 0) return null
  if (formValues == null) {
    return oneOf.find((b) => b.type === 'null') ?? null
  }
  return oneOf.find((b) => b.type === 'object' && b.properties) ?? null
}

function firstEnumValue(schema: JSONSchema7): unknown {
  return Array.isArray(schema.enum) && schema.enum.length > 0
    ? schema.enum[0]
    : undefined
}

/** JSONSchema7 allows items to be a single schema, an array (tuple), or boolean. */
function getArrayItemSchema(
  s: JSONSchema7,
  root: JSONSchema7,
  index: number,
): JSONSchema7 {
  const items = (s as { items?: JSONSchema7 | JSONSchema7[] | boolean }).items
  if (items === undefined || items === null) return {}
  if (typeof items === 'boolean') {
    return {} // true/false: no per-item schema for normalization
  }
  if (Array.isArray(items)) {
    // Tuple: use schema at index, or last schema for extra items
    const raw = items[index] ?? items[items.length - 1]
    const schema =
      raw !== undefined &&
      typeof raw === 'object' &&
      raw !== null &&
      !Array.isArray(raw)
        ? (raw as JSONSchema7)
        : {}
    return getSchema(schema, root)
  }
  return getSchema(items as JSONSchema7, root)
}

export function convertRjsfInternalFormats(
  schema: JSONSchema7,
  formValues: FormState['formData'],
  rootSchema?: JSONSchema7,
): FormState['formData'] {
  // When no root is provided, normalize so $defs are at root (resolveRef looks at root.$defs).
  // SchemaForm already runs normalizeSchemaDefs before render; this handles raw/caller schemas.
  const root = rootSchema ?? normalizeSchemaDefs(schema)
  const s = getSchema(schema, root)

  // Leaf: duration string → human-readable format
  if (
    s.type === 'string' &&
    s.format === 'duration' &&
    (isString(formValues) || isNumber(formValues))
  ) {
    const valueStr = formValues.toString() ?? ''
    return formatDuration(
      dayjs.duration(
        getMillisecondsFromDurationString(valueStr)!,
        'milliseconds',
      ),
    )
  }

  // Leaf: null/undefined at enum → first enum value (matches dropdown display)
  const enumDefault = firstEnumValue(s)
  if (
    (formValues === null || formValues === undefined) &&
    enumDefault !== undefined
  ) {
    return enumDefault as FormState['formData']
  }

  // Recursive: array → normalize each item (items may be single schema, tuple, or boolean)
  if (s.type === 'array' && Array.isArray(formValues)) {
    return formValues.map((item, index) =>
      convertRjsfInternalFormats(
        getArrayItemSchema(s, root, index),
        item,
        root,
      ),
    ) as FormState['formData']
  }

  // Recursive: oneOf (e.g. optional X | null) → recurse into the matching branch
  if (!s?.properties) {
    const branch = getOneOfBranch(s, formValues)
    if (branch?.properties && isObject(formValues)) {
      return convertRjsfInternalFormats(branch, formValues, root)
    }
    return formValues
  }

  // Recursive: object → normalize each property
  const valuesToUpdate = Object.entries(s.properties).reduce(
    (acc, [key, schemaEntry]) => {
      if (schemaEntry === true || schemaEntry === false) return acc
      const entrySchema = schemaEntry as JSONSchema7
      if (!(key in formValues)) return acc
      let raw = (formValues as Record<string, unknown>)[key]
      const resolved = getSchema(entrySchema, root)
      if (resolved.type === 'array') {
        raw = ensureArrayIfNumericKeys(raw)
      }
      acc[key] = convertRjsfInternalFormats(entrySchema, raw, root)
      return acc
    },
    {} as FormState['formData'],
  )
  return { ...formValues, ...valuesToUpdate }
}

/**
 * Normalize form data: durations, enum nulls → first enum value, and $ref resolution.
 * Safe to call on every RJSF change and again at submit.
 */
export function normalizeFormData(
  schema: JSONSchema7 | undefined,
  data: FormState['formData'],
): FormState['formData'] {
  if (!schema) return data
  return convertRjsfInternalFormats(schema, data)
}

const FORMAT_FALLBACK_MAPPING: Record<string, string> = {
  'alt-date': 'date',
  'alt-datetime': 'datetime',
}

/*
 * Provide fallback of some unsupported formats to formats that we have implementation for
 */
export function getSchemaFormatFallback(
  orgFormat?: string,
): string | undefined {
  return orgFormat && FORMAT_FALLBACK_MAPPING[orgFormat]
    ? FORMAT_FALLBACK_MAPPING[orgFormat]
    : orgFormat
}

export function fromFormattedValueToText(
  rawValue: string,
  format: string | undefined,
): string {
  switch (getSchemaFormatFallback(format)) {
    case 'date':
      return dayjs.utc(rawValue).format('YYYY-MM-DD')
    case 'datetime':
      return dayjs.utc(rawValue).format('YYYY-MM-DD HH:mm:ss')
    case 'time':
      return dayjs.utc(rawValue).format('HH:mm:ss')
    default:
      return rawValue
  }
}

export function fromTextToFormattedValue(
  textValue: string,
  schemaFormat: string | undefined,
): string {
  switch (getSchemaFormatFallback(schemaFormat)) {
    case 'date':
    case 'datetime':
      try {
        return dayjs.utc(textValue).toISOString()
      } catch (e) {
        console.warn(e)
        return textValue
      }
    case 'time':
      try {
        return dayjs(textValue, 'HH:mm[:ss]').format('HH:mm:ss')
      } catch (e) {
        console.warn(e)
        return textValue
      }

    default:
      return textValue
  }
}

/**
 * Gets a human-readable title for a schema option
 */
function getOptionTitle(
  option: JSONSchema7 | Record<string, unknown>,
  defs: Record<string, JSONSchema7 | Record<string, unknown>>,
): string {
  // If it already has a title, use it
  if (option.title && typeof option.title === 'string') {
    return option.title
  }

  // If it's a $ref, try to get the title from the referenced definition
  if ('$ref' in option && option.$ref && typeof option.$ref === 'string') {
    const refName = option.$ref.split('/').pop()
    if (refName && defs[refName]) {
      const refDef = defs[refName]
      if (refDef && typeof refDef === 'object' && 'title' in refDef) {
        if (refDef.title && typeof refDef.title === 'string') {
          return refDef.title
        }
      }
      // Fallback to the definition name itself
      return refName
    }
  }

  // For null type, return "None"
  if ('type' in option && option.type === 'null') {
    return 'None'
  }

  // For simple types, capitalize them
  if ('type' in option && option.type && typeof option.type === 'string') {
    return option.type.charAt(0).toUpperCase() + option.type.slice(1)
  }

  // Fallback
  return 'Unknown'
}

/**
 * Normalizes a JSON schema by moving nested $defs to the root level
 * and adding titles to anyOf/oneOf options that don't have them.
 * JSON Schema requires $defs to be at the root, but sometimes schemas
 * have $defs nested inside properties, which causes $ref resolution to fail.
 */
export function normalizeSchemaDefs(schema: JSONSchema7): JSONSchema7 {
  if (!schema || typeof schema !== 'object') {
    return schema
  }

  const processedSchema = cloneDeep(schema)
  const rootDefs: Record<string, JSONSchema7 | Record<string, unknown>> = {}

  // Initialize rootDefs from processedSchema.$defs, filtering out boolean values
  if (processedSchema.$defs && typeof processedSchema.$defs === 'object') {
    Object.entries(processedSchema.$defs).forEach(([key, value]) => {
      if (value && typeof value === 'object' && !Array.isArray(value)) {
        rootDefs[key] = value as JSONSchema7 | Record<string, unknown>
      }
    })
  }

  /**
   * Recursively finds and collects all $defs from nested objects
   * @param obj - The object to process
   * @param isRoot - Whether this is the root schema object (to preserve root $defs)
   */
  function collectDefs(
    obj: JSONSchema7 | Record<string, unknown> | unknown[],
    isRoot = false,
  ): void {
    if (!obj || typeof obj !== 'object') {
      return
    }

    if (Array.isArray(obj)) {
      obj.forEach((item) => {
        if (item && typeof item === 'object') {
          collectDefs(item as Record<string, unknown>, false)
        }
      })
      return
    }

    const objRecord = obj as Record<string, unknown>

    // If this object has $defs (and it's not the root), merge them into rootDefs
    if (!isRoot && objRecord.$defs && typeof objRecord.$defs === 'object') {
      const defs = objRecord.$defs as Record<string, unknown>
      Object.entries(defs).forEach(([key, value]) => {
        // Merge nested definitions into root, keeping existing root definitions
        if (
          !rootDefs[key] &&
          value &&
          typeof value === 'object' &&
          !Array.isArray(value)
        ) {
          rootDefs[key] = value as JSONSchema7 | Record<string, unknown>
          // Recursively process nested definitions for nested $defs
          collectDefs(value as Record<string, unknown>, false)
        }
      })
      // Remove $defs from nested location
      delete objRecord.$defs
    }

    // Recursively process all properties
    Object.entries(objRecord).forEach(([_key, value]) => {
      // Skip $defs as we already processed it (or it's the root which we preserve)
      if (_key === '$defs') {
        return
      }
      if (value && typeof value === 'object') {
        collectDefs(
          value as JSONSchema7 | Record<string, unknown> | unknown[],
          false,
        )
      }
    })
  }

  // Collect all nested $defs (pass true to indicate this is the root)
  collectDefs(processedSchema, true)

  // Set root-level $defs (merge any collected nested defs with existing root defs)
  if (Object.keys(rootDefs).length > 0) {
    processedSchema.$defs = rootDefs
  }

  // Now normalize anyOf/oneOf options with the complete rootDefs
  function normalizeOptions(
    obj: JSONSchema7 | Record<string, unknown> | unknown[],
  ): void {
    if (!obj || typeof obj !== 'object') {
      return
    }

    if (Array.isArray(obj)) {
      obj.forEach((item) => {
        normalizeOptions(item as JSONSchema7 | Record<string, unknown>)
      })
      return
    }

    const objRecord = obj as Record<string, unknown>

    if (Array.isArray(objRecord.anyOf)) {
      objRecord.anyOf = objRecord.anyOf.map((option: unknown) => {
        const optionObj = option as JSONSchema7 | Record<string, unknown>
        if (!('title' in optionObj) || !optionObj.title) {
          return {
            ...optionObj,
            title: getOptionTitle(optionObj, rootDefs),
          }
        }
        return option
      })
    }
    if (Array.isArray(objRecord.oneOf)) {
      objRecord.oneOf = objRecord.oneOf.map((option: unknown) => {
        const optionObj = option as JSONSchema7 | Record<string, unknown>
        if (!('title' in optionObj) || !optionObj.title) {
          return {
            ...optionObj,
            title: getOptionTitle(optionObj, rootDefs),
          }
        }
        return option
      })
    }

    // Recursively process all properties
    Object.entries(objRecord).forEach(([_key, value]) => {
      if (value && typeof value === 'object') {
        normalizeOptions(value as JSONSchema7 | Record<string, unknown>)
      }
    })
  }

  // Normalize options with complete rootDefs
  normalizeOptions(processedSchema)

  return processedSchema
}

export const skipUndefinedFormValues = (
  formValues: FormState['formData'],
): FormState['formData'] => {
  if (Array.isArray(formValues)) {
    return formValues.map((item) =>
      skipUndefinedFormValues(item),
    ) as FormState['formData']
  }
  if (!isObject(formValues)) {
    return formValues
  }
  return Object.entries(formValues).reduce(
    (newValues, [key, val]) => {
      if (val !== undefined) {
        newValues[key] = skipUndefinedFormValues(val)
      }
      return newValues
    },
    {} as FormState['formData'],
  )
}
