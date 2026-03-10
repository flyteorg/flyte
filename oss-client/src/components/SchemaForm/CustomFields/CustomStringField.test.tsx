import React from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { JSONSchema7 } from 'json-schema'
import { Registry } from '@rjsf/utils'
import CustomStringField from './CustomStringField'

const mockParams = {
  onChange: vi.fn(),
  onBlur: vi.fn(),
  onFocus: vi.fn(),
  name: 'mock',
  idSchema: {
    $id: '',
  },
  registry: {} as Registry,
}

const dateSchema: JSONSchema7 = {
  type: 'string',
  format: 'date',
  default: '2025-12-12',
}
const timeSchema: JSONSchema7 = {
  type: 'string',
  format: 'time',
  default: '22:12:01',
}
const datetimeSchema: JSONSchema7 = {
  type: 'string',
  format: 'datetime',
  default: '2025-12-12 12:12:12',
}

const textSchema: JSONSchema7 = {
  type: 'string',
  default: 'lorem ipsum',
}

describe('CustomStringField', () => {
  it('should handle date format', () => {
    render(<CustomStringField {...mockParams} schema={dateSchema} />)

    expect(screen.getByTestId('input-mock')).toHaveAttribute('type', 'date')

    fireEvent.change(screen.getByTestId('input-mock'), {
      target: { value: '2025-10-10' },
    })

    expect(mockParams.onChange).toHaveBeenCalled()
    expect(mockParams.onChange).toHaveBeenCalledWith('2025-10-10T00:00:00.000Z')
  })

  it('should handle time format', () => {
    render(<CustomStringField {...mockParams} schema={timeSchema} />)

    expect(screen.getByTestId('input-mock')).toHaveAttribute('type', 'time')

    fireEvent.change(screen.getByTestId('input-mock'), {
      target: { value: '22:22' },
    })

    expect(mockParams.onChange).toHaveBeenCalled()
    expect(mockParams.onChange).toHaveBeenCalledWith('22:22:00')
  })

  it('should handle datetime format', () => {
    render(<CustomStringField {...mockParams} schema={datetimeSchema} />)

    expect(screen.getByTestId('input-mock')).toHaveAttribute(
      'type',
      'datetime-local',
    )

    fireEvent.change(screen.getByTestId('input-mock'), {
      target: { value: '2025-11-11 22:22:22' },
    })

    expect(mockParams.onChange).toHaveBeenCalled()
    expect(mockParams.onChange).toHaveBeenCalledWith('2025-11-11T22:22:22.000Z')
  })

  it('falls back to date/datetime for undocumented RJSF alt formats', () => {
    render(
      <CustomStringField
        {...mockParams}
        name="altdatetime"
        schema={{ ...datetimeSchema, format: 'alt-datetime' }}
      />,
    )
    expect(screen.getByTestId('input-altdatetime')).toHaveAttribute(
      'type',
      'datetime-local',
    )

    render(
      <CustomStringField
        {...mockParams}
        name="altdate"
        schema={{ ...dateSchema, format: 'alt-date' }}
      />,
    )
    expect(screen.getByTestId('input-altdate')).toHaveAttribute('type', 'date')
  })

  it('uses text input for short strings', () => {
    const shortText = 'short'
    render(
      <CustomStringField
        {...mockParams}
        schema={textSchema}
        formData={shortText}
      />,
    )
    expect(screen.getByTestId('input-mock')).toBeInstanceOf(HTMLInputElement)
    expect(screen.getByTestId('input-mock')).toHaveAttribute('type', 'text')
  })

  it('uses textarea for long strings', () => {
    const longText = 'lorem ipsum dolor sit amet'.repeat(10)
    render(
      <CustomStringField
        {...mockParams}
        schema={textSchema}
        formData={longText}
      />,
    )

    expect(screen.getByTestId('input-mock')).toBeInstanceOf(HTMLTextAreaElement)
  })
})
