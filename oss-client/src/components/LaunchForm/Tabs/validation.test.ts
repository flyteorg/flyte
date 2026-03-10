import { describe, it, expect } from 'vitest'
import { getKvPairsValidationError } from './validation'

describe('getKvPairsValidationError', () => {
  it('returns null for undefined or empty array', () => {
    expect(getKvPairsValidationError(undefined)).toBeNull()
    expect(getKvPairsValidationError([])).toBeNull()
  })

  it('returns null for single empty key-value pair (default row)', () => {
    expect(getKvPairsValidationError([{ key: '', value: '' }])).toBeNull()
    expect(getKvPairsValidationError([{ key: '   ', value: '' }])).toBeNull()
  })

  it('returns "Key is required" when a pair has value but empty key', () => {
    expect(getKvPairsValidationError([{ key: '', value: 'some-value' }])).toBe(
      'Key is required',
    )
    expect(getKvPairsValidationError([{ key: '   ', value: 'v' }])).toBe(
      'Key is required',
    )
    expect(
      getKvPairsValidationError([
        { key: 'valid', value: 'v1' },
        { key: '', value: 'v2' },
      ]),
    ).toBe('Key is required')
  })

  it('returns null when all pairs have keys and values', () => {
    expect(
      getKvPairsValidationError([
        { key: 'a', value: '1' },
        { key: 'b', value: '2' },
      ]),
    ).toBeNull()
  })

  it('returns error for keys containing spaces', () => {
    expect(
      getKvPairsValidationError([{ key: 'key with spaces', value: 'v' }]),
    ).toBe('Keys cannot contain spaces: key with spaces')
    expect(
      getKvPairsValidationError([
        { key: 'a b', value: '1' },
        { key: 'c d', value: '2' },
      ]),
    ).toContain('Keys cannot contain spaces')
    expect(
      getKvPairsValidationError([
        { key: 'a b', value: '1' },
        { key: 'c d', value: '2' },
      ]),
    ).toContain('a b')
    expect(
      getKvPairsValidationError([
        { key: 'a b', value: '1' },
        { key: 'c d', value: '2' },
      ]),
    ).toContain('c d')
  })

  it('returns error for duplicate keys', () => {
    expect(
      getKvPairsValidationError([
        { key: 'same', value: '1' },
        { key: 'same', value: '2' },
      ]),
    ).toBe('Duplicate keys found: same')
    expect(
      getKvPairsValidationError([
        { key: 'a', value: '1' },
        { key: 'b', value: '2' },
        { key: 'a', value: '3' },
      ]),
    ).toContain('Duplicate keys found')
    expect(
      getKvPairsValidationError([
        { key: 'a', value: '1' },
        { key: 'b', value: '2' },
        { key: 'a', value: '3' },
      ]),
    ).toContain('a')
  })

  it('prioritizes value-with-empty-key over spaces and duplicates', () => {
    expect(getKvPairsValidationError([{ key: '', value: 'x' }])).toBe(
      'Key is required',
    )
    expect(
      getKvPairsValidationError([
        { key: 'a b', value: '1' },
        { key: '', value: '2' },
      ]),
    ).toBe('Key is required')
  })

  it('prioritizes spaces over duplicates when no empty-key-with-value', () => {
    expect(
      getKvPairsValidationError([
        { key: 'a b', value: '1' },
        { key: 'a b', value: '2' },
      ]),
    ).toContain('Keys cannot contain spaces')
  })

  it('treats value-with-empty-key when key is undefined', () => {
    expect(getKvPairsValidationError([{ value: 'x' }])).toBe('Key is required')
    expect(getKvPairsValidationError([{ key: undefined, value: 'x' }])).toBe(
      'Key is required',
    )
  })

  it('does not error when value is only whitespace (no real value)', () => {
    expect(getKvPairsValidationError([{ key: '', value: '   ' }])).toBeNull()
  })

  it('duplicate keys are detected after trimming', () => {
    expect(
      getKvPairsValidationError([
        { key: 'a', value: '1' },
        { key: '  a  ', value: '2' },
      ]),
    ).toBe('Duplicate keys found: a')
  })

  it('multiple duplicate keys appear in error message', () => {
    const result = getKvPairsValidationError([
      { key: 'x', value: '1' },
      { key: 'x', value: '2' },
      { key: 'y', value: '3' },
      { key: 'y', value: '4' },
    ])
    expect(result).toContain('Duplicate keys found')
    expect(result).toContain('x')
    expect(result).toContain('y')
  })

  it('allows empty value when key is present', () => {
    expect(
      getKvPairsValidationError([{ key: 'valid-key', value: '' }]),
    ).toBeNull()
  })
})
