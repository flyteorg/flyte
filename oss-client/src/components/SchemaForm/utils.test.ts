import { describe, it, expect } from 'vitest'
import {
  convertRjsfInternalFormats,
  getFormDataFromSchemaDefaults,
  normalizeNoneTypeToNull,
  normalizeSchemaDefs,
  skipUndefinedFormValues,
} from '@/components/SchemaForm/utils'
import type { JSONSchema7 } from 'json-schema'

describe('getFormDataFromSchemaDefaults', () => {
  it('returns default value', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'string',
      default: 'test',
    })

    expect(result).toBe('test')
  })
  it('returns nested properties', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'object',
      properties: {
        prop1: { type: 'string', default: 'string' },
        prop2: { type: 'number', default: 123 },
      },
    })

    expect(result).toStrictEqual({
      prop1: 'string',
      prop2: 123,
    })
  })
  it('handles form with no properties', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'object',
      title: 'title',
    })

    expect(result).toStrictEqual({})
  })
  it('handles properties with no defaults', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'object',
      properties: {
        prop1: { type: 'string', default: 'string' },
        prop2: { type: 'number', default: 123 },
        prop3: { type: 'string' },
      },
    })

    expect(result).toStrictEqual({
      prop1: 'string',
      prop2: 123,
      prop3: undefined,
    })
  })
  it('handles nested objects', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'object',
      properties: {
        obj: {
          type: 'object',
          properties: {
            prop1: { type: 'string', default: 'string' },
          },
        },
        obj2: {
          type: 'object',
          properties: {
            prop2: { type: 'string' },
          },
        },
        obj3: {
          type: 'object',
          properties: {
            lvl3Obj: {
              type: 'object',
              properties: {
                deepProp1: { type: 'string', default: 'test' },
              },
            },
          },
        },
      },
    })

    expect(result).toStrictEqual({
      obj: { prop1: 'string' },
      obj2: { prop2: undefined },
      obj3: { lvl3Obj: { deepProp1: 'test' } },
    })
  })
  it('converts NONETYPE default { type: "null" } to null', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'object',
      properties: {
        custom_filtering_model_s3_path: {
          type: 'string',
          default: { type: 'null' },
        },
        other: { type: 'string', default: 'value' },
      },
    })
    expect(result).toStrictEqual({
      custom_filtering_model_s3_path: null,
      other: 'value',
    })
  })
  it('defaults enum without default to first enum value', () => {
    const result = getFormDataFromSchemaDefaults({
      type: 'object',
      properties: {
        choice: { type: 'string', enum: ['A', 'B', 'C'] },
        withDefault: { type: 'string', enum: ['X', 'Y'], default: 'Y' },
      },
    })
    expect(result).toStrictEqual({ choice: 'A', withDefault: 'Y' })
  })
})

describe('normalizeNoneTypeToNull', () => {
  it('replaces { type: "null" } with null', () => {
    expect(normalizeNoneTypeToNull({ type: 'null' })).toBe(null)
  })
  it('recursively normalizes nested objects', () => {
    expect(
      normalizeNoneTypeToNull({
        a: { type: 'null' },
        b: 'keep',
        c: { inner: { type: 'null' } },
      }),
    ).toStrictEqual({ a: null, b: 'keep', c: { inner: null } })
  })
  it('leaves non-none-type values unchanged', () => {
    expect(normalizeNoneTypeToNull('s')).toBe('s')
    expect(normalizeNoneTypeToNull(1)).toBe(1)
    expect(normalizeNoneTypeToNull(null)).toBe(null)
    expect(normalizeNoneTypeToNull({ type: 'string' })).toStrictEqual({
      type: 'string',
    })
  })
})

describe('normalizeSchemaDefs', () => {
  it('moves nested $defs to root level', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch_info: {
          type: 'object',
          $defs: {
            DataFrame: {
              type: 'object',
              title: 'DataFrame',
              properties: {
                uri: { type: 'string' },
                format: { type: 'string' },
              },
            },
            ExtractionBatch: {
              type: 'object',
              title: 'ExtractionBatch',
              properties: {
                batch: { $ref: '#/$defs/DataFrame' },
              },
            },
          },
          properties: {
            batch: {
              $ref: '#/$defs/ExtractionBatch',
            },
          },
        },
      },
    }

    const result = normalizeSchemaDefs(schema)

    expect(result.$defs).toBeDefined()
    expect(result.$defs?.DataFrame).toBeDefined()
    expect(result.$defs?.ExtractionBatch).toBeDefined()
    expect((result.properties?.batch_info as JSONSchema7).$defs).toBeUndefined()
  })

  it('adds title to null type in anyOf', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch: {
          anyOf: [{ $ref: '#/$defs/DataFrame' }, { type: 'null' }],
        },
      },
      $defs: {
        DataFrame: {
          type: 'object',
          title: 'DataFrame',
        },
      },
    }

    const result = normalizeSchemaDefs(schema)
    const batchProp = result.properties?.batch as JSONSchema7
    const anyOf = batchProp.anyOf as JSONSchema7[]

    expect(anyOf).toBeDefined()
    expect(anyOf.length).toBe(2)
    const nullOption = anyOf.find((opt) => opt.type === 'null')
    expect(nullOption?.title).toBe('None')
  })

  it('adds title to null type in oneOf', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch: {
          oneOf: [{ $ref: '#/$defs/DataFrame' }, { type: 'null' }],
        },
      },
      $defs: {
        DataFrame: {
          type: 'object',
          title: 'DataFrame',
        },
      },
    }

    const result = normalizeSchemaDefs(schema)
    const batchProp = result.properties?.batch as JSONSchema7
    const oneOf = batchProp.oneOf as JSONSchema7[]

    expect(oneOf).toBeDefined()
    expect(oneOf.length).toBe(2)
    const nullOption = oneOf.find((opt) => opt.type === 'null')
    expect(nullOption?.title).toBe('None')
  })

  it('extracts title from $ref to referenced definition', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch: {
          anyOf: [{ $ref: '#/$defs/DataFrame' }, { type: 'null' }],
        },
      },
      $defs: {
        DataFrame: {
          type: 'object',
          title: 'DataFrame',
        },
      },
    }

    const result = normalizeSchemaDefs(schema)
    const batchProp = result.properties?.batch as JSONSchema7
    const anyOf = batchProp.anyOf as JSONSchema7[]

    const refOption = anyOf.find((opt) => opt.$ref)
    expect(refOption?.title).toBe('DataFrame')
  })

  it('preserves existing titles in anyOf/oneOf', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch: {
          anyOf: [
            { $ref: '#/$defs/DataFrame', title: 'Custom Title' },
            { type: 'null' },
          ],
        },
      },
      $defs: {
        DataFrame: {
          type: 'object',
          title: 'DataFrame',
        },
      },
    }

    const result = normalizeSchemaDefs(schema)
    const batchProp = result.properties?.batch as JSONSchema7
    const anyOf = batchProp.anyOf as JSONSchema7[]

    const refOption = anyOf.find((opt) => opt.$ref)
    expect(refOption?.title).toBe('Custom Title')
  })

  it('handles nested anyOf/oneOf in properties', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch_info: {
          type: 'object',
          properties: {
            batch: {
              anyOf: [{ $ref: '#/$defs/DataFrame' }, { type: 'null' }],
            },
          },
        },
      },
      $defs: {
        DataFrame: {
          type: 'object',
          title: 'DataFrame',
        },
      },
    }

    const result = normalizeSchemaDefs(schema)
    const batchInfoProp = result.properties?.batch_info as JSONSchema7
    const batchProp = (batchInfoProp.properties as Record<string, JSONSchema7>)
      ?.batch
    const anyOf = batchProp.anyOf as JSONSchema7[]

    expect(anyOf).toBeDefined()
    const nullOption = anyOf.find((opt) => opt.type === 'null')
    expect(nullOption?.title).toBe('None')
  })

  it('handles schema with no $defs', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        name: { type: 'string' },
      },
    }

    const result = normalizeSchemaDefs(schema)

    expect(result).toEqual(schema)
    expect(result.$defs).toBeUndefined()
  })

  it('handles schema with no anyOf/oneOf', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        name: { type: 'string' },
      },
      $defs: {
        MyType: {
          type: 'object',
          title: 'MyType',
        },
      },
    }

    const result = normalizeSchemaDefs(schema)

    expect(result.$defs).toBeDefined()
    expect(result.$defs?.MyType).toBeDefined()
  })

  it('handles empty schema', () => {
    const schema: JSONSchema7 = {}

    const result = normalizeSchemaDefs(schema)

    expect(result).toEqual({})
  })

  it('handles null/undefined input', () => {
    expect(normalizeSchemaDefs(null as unknown as JSONSchema7)).toBeNull()
    expect(
      normalizeSchemaDefs(undefined as unknown as JSONSchema7),
    ).toBeUndefined()
  })

  it('merges nested $defs without overwriting root $defs', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      $defs: {
        RootType: {
          type: 'object',
          title: 'RootType',
        },
      },
      properties: {
        nested: {
          type: 'object',
          $defs: {
            NestedType: {
              type: 'object',
              title: 'NestedType',
            },
          },
        },
      },
    }

    const result = normalizeSchemaDefs(schema)

    expect(result.$defs?.RootType).toBeDefined()
    expect(result.$defs?.NestedType).toBeDefined()
    expect((result.properties?.nested as JSONSchema7).$defs).toBeUndefined()
  })

  it('capitalizes simple types without titles', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        field: {
          anyOf: [{ type: 'string' }, { type: 'number' }, { type: 'null' }],
        },
      },
    }

    const result = normalizeSchemaDefs(schema)
    const fieldProp = result.properties?.field as JSONSchema7
    const anyOf = fieldProp.anyOf as JSONSchema7[]

    const stringOption = anyOf.find((opt) => opt.type === 'string')
    const numberOption = anyOf.find((opt) => opt.type === 'number')
    const nullOption = anyOf.find((opt) => opt.type === 'null')

    expect(stringOption?.title).toBe('String')
    expect(numberOption?.title).toBe('Number')
    expect(nullOption?.title).toBe('None')
  })

  it('handles complex nested structure with $defs and anyOf', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        batch_info: {
          type: 'object',
          $defs: {
            DataFrame: {
              type: 'object',
              title: 'DataFrame',
              properties: {
                uri: { type: 'string' },
                format: { type: 'string' },
              },
            },
            ExtractionBatch: {
              type: 'object',
              title: 'ExtractionBatch',
              properties: {
                batch: {
                  anyOf: [{ $ref: '#/$defs/DataFrame' }, { type: 'null' }],
                },
                report: { $ref: '#/$defs/DataFrame' },
                model: { type: 'string' },
              },
            },
          },
          properties: {
            batch: {
              $ref: '#/$defs/ExtractionBatch',
            },
          },
        },
      },
    }

    const result = normalizeSchemaDefs(schema)

    // Check $defs moved to root
    expect(result.$defs?.DataFrame).toBeDefined()
    expect(result.$defs?.ExtractionBatch).toBeDefined()

    // Check nested $defs removed
    const batchInfo = result.properties?.batch_info as JSONSchema7
    expect(batchInfo.$defs).toBeUndefined()

    // Check anyOf titles added
    const extractionBatch = result.$defs?.ExtractionBatch as JSONSchema7
    const batchProp = (
      extractionBatch.properties as Record<string, JSONSchema7>
    )?.batch
    const anyOf = batchProp.anyOf as JSONSchema7[]
    expect(anyOf).toBeDefined()
    const nullOption = anyOf.find((opt) => opt.type === 'null')
    expect(nullOption?.title).toBe('None')
    const refOption = anyOf.find((opt) => opt.$ref)
    expect(refOption?.title).toBe('DataFrame')
  })
})

describe('convertRjsfInternalFormats', () => {
  const durationSchema: JSONSchema7 = {
    type: 'string',
    format: 'duration',
    default: '4h',
  }
  it('converts duration from ISO8601', () => {
    const result = convertRjsfInternalFormats(durationSchema, 'PT2H')
    expect(result).toBe('2h')
  })
  it('leaves properly formatted value unchanged', () => {
    const result = convertRjsfInternalFormats(durationSchema, '2h')
    expect(result).toBe('2h')
  })
  it('changes value of nested duration property', () => {
    const result = convertRjsfInternalFormats(
      {
        type: 'object',
        properties: {
          durationLvl1: durationSchema,
          prop2: {
            type: 'object',
            properties: {
              durationLvl2: durationSchema,
              otherProp: { type: 'string', format: 'datetime' },
              nestedObj: {
                type: 'object',
                properties: {
                  durationLvl3: durationSchema,
                },
              },
            },
          },
        },
      },
      {
        durationLvl1: 'PT2H30M',
        prop2: {
          durationLvl2: 'PT3H59M',
          otherProp: '2025-01-01 23:59:59',
          nestedObj: { durationLvl3: 'PT1H' },
        },
      },
    )
    expect(result).toStrictEqual({
      durationLvl1: '2h30m',
      prop2: {
        durationLvl2: '3h59m',
        otherProp: '2025-01-01 23:59:59',
        nestedObj: { durationLvl3: '1h' },
      },
    })
  })

  it('supports raw miliseconds', () => {
    const result = convertRjsfInternalFormats(durationSchema, 3600_000)
    expect(result).toBe('1h')
  })

  it('normalizes object-with-numeric-keys to array when schema type is array', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        items: { type: 'array', items: { type: 'string' } },
      },
    }
    const formValues = { items: { '0': 'a', '1': 'b', '2': 'c' } }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({ items: ['a', 'b', 'c'] })
  })

  it('replaces null in array-of-enum with first enum value', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        datasets: {
          type: 'array',
          items: {
            type: 'string',
            enum: ['S2_MED_PLANTING', 'S2_MED_HARVEST', 'OTHER'],
          },
        },
      },
    }
    const formValues = { datasets: ['S2_MED_PLANTING', null] }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({
      datasets: ['S2_MED_PLANTING', 'S2_MED_PLANTING'],
    })
  })

  it('replaces null for any nested enum (e.g. inside object)', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        level1: {
          type: 'object',
          properties: {
            choice: { type: 'string', enum: ['X', 'Y', 'Z'] },
          },
        },
      },
    }
    const formValues = { level1: { choice: null } }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({ level1: { choice: 'X' } })
  })

  it('replaces null in array-of-enum when items use $ref', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        datasets: {
          type: 'array',
          items: { $ref: '#/$defs/DatasetEnum' },
        },
      },
      $defs: {
        DatasetEnum: {
          type: 'string',
          enum: ['S2_MED_PLANTING', 'S2_MED_HARVEST', 'OTHER'],
        },
      },
    }
    const formValues = { datasets: ['S2_MED_PLANTING', null] }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({
      datasets: ['S2_MED_PLANTING', 'S2_MED_PLANTING'],
    })
  })

  it('resolves $ref when schema has nested $defs (normalizes at entry)', () => {
    // Schema with $defs nested under a property (not at root). Without normalizing at entry,
    // root.$defs would be undefined and enum-null replacement would not apply.
    const schemaWithNestedDefs = {
      type: 'object',
      properties: {
        choice: {
          type: 'array',
          items: { $ref: '#/$defs/MyEnum' },
          $defs: {
            MyEnum: { type: 'string', enum: ['X', 'Y', 'Z'] },
          },
        },
      },
    } as JSONSchema7
    const formValues = { choice: [null] }
    const result = convertRjsfInternalFormats(schemaWithNestedDefs, formValues)
    expect(result).toStrictEqual({ choice: ['X'] })
  })

  it('handles tuple items: each index uses its schema for enum default', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        pair: {
          type: 'array',
          items: [
            { type: 'string', enum: ['A', 'B'] },
            { type: 'string', enum: ['X', 'Y'] },
          ],
        },
      },
    }
    const formValues = { pair: [null, null] }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({ pair: ['A', 'X'] })
  })

  it('handles boolean items: passthrough (no enum normalization)', () => {
    const schema: JSONSchema7 = {
      type: 'object',
      properties: {
        any: { type: 'array', items: true },
      },
    }
    const formValues = { any: ['foo', null] }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({ any: ['foo', null] })
  })

  it('recurses into oneOf object branch and normalizes nested array fields', () => {
    const schema: JSONSchema7 = {
      oneOf: [
        {
          type: 'object',
          properties: {
            document_extensions: { type: 'array', items: { type: 'string' } },
            image_extensions: { type: 'array', items: { type: 'string' } },
          },
        },
        { type: 'null' },
      ],
    }
    const formValues = {
      document_extensions: { '0': 'pdf', '1': 'djvu' },
      image_extensions: { '0': 'png', '1': 'jpg' },
    }
    const result = convertRjsfInternalFormats(schema, formValues)
    expect(result).toStrictEqual({
      document_extensions: ['pdf', 'djvu'],
      image_extensions: ['png', 'jpg'],
    })
  })
})

describe('skipUndefinedFormValues', () => {
  it('skips undefined form values', () => {
    const result = skipUndefinedFormValues({
      prop1: 'val1',
      prop2: undefined,
      prop3: 'val3',
    })
    expect(result).toStrictEqual({ prop1: 'val1', prop3: 'val3' })
  })

  it('skips undefined values in nested objects', () => {
    const result = skipUndefinedFormValues({
      prop1: 'val1',
      prop2: {
        sub1: 'subval1',
        sub2: undefined,
        sub3: 'subval3',
      },
    })
    expect(result.prop2).toStrictEqual({ sub1: 'subval1', sub3: 'subval3' })
  })

  it('leaves non-object values intact', () => {
    expect(skipUndefinedFormValues(1234)).toBe(1234)
    expect(skipUndefinedFormValues(null)).toBe(null)
    expect(skipUndefinedFormValues(undefined)).toBe(undefined)
    expect(skipUndefinedFormValues(0)).toBe(0)
    expect(skipUndefinedFormValues(false)).toBe(false)
  })

  it('preserves arrays (does not turn them into objects with numeric keys)', () => {
    const result = skipUndefinedFormValues({
      extensions: ['pdf', 'djvu'],
      nested: { list: [1, 2, 3] },
    })
    expect(result).toStrictEqual({
      extensions: ['pdf', 'djvu'],
      nested: { list: [1, 2, 3] },
    })
    expect(Array.isArray((result as Record<string, unknown>).extensions)).toBe(
      true,
    )
    expect(
      Array.isArray((result as Record<string, unknown>).nested?.list),
    ).toBe(true)
  })
})
