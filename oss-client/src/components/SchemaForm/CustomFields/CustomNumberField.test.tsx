import React from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import CustomNumberField from './CustomNumberField'
import { JSONSchema7 } from 'json-schema'
import { Registry } from '@rjsf/utils'

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

const integerSchema: JSONSchema7 = { default: 123, type: 'integer' }
const floatSchema: JSONSchema7 = { default: 12.3, type: 'number' }

describe('CustomNumberField', () => {
  it('renders integer crashing', () => {
    render(<CustomNumberField schema={integerSchema} {...mockParams} />)

    expect(screen.getByTestId('input-mock')).toBeInTheDocument()
    expect(screen.getByTestId('input-mock')).toHaveAttribute('value', integerSchema.default?.toString())
  })
  it('renders float crashing', () => {
    render(<CustomNumberField schema={floatSchema} {...mockParams} />)

    expect(screen.getByTestId('input-mock')).toBeInTheDocument()
    expect(screen.getByTestId('input-mock')).toHaveAttribute('value', floatSchema.default?.toString())
  })
  it('renders with BigInt value', () => {
    const bigIntStr = `${Number.MAX_SAFE_INTEGER}123`
    render(<CustomNumberField schema={integerSchema} {...mockParams} formData={BigInt(bigIntStr)} />)

    expect(screen.getByTestId('input-mock')).toBeInTheDocument()
    expect(screen.getByTestId('input-mock')).toHaveAttribute('value', bigIntStr)

  })

  it('executes event handlers', () => {
    render(<CustomNumberField schema={integerSchema} {...mockParams} />)

    fireEvent.focusIn(screen.getByTestId('input-mock'))
    expect(mockParams.onFocus).toHaveBeenCalled()

    fireEvent.blur(screen.getByTestId('input-mock'))
    expect(mockParams.onBlur).toHaveBeenCalled()

    fireEvent.change(screen.getByTestId('input-mock'),{target: {value: '234'}})
    expect(mockParams.onChange).toHaveBeenCalled()
    expect(mockParams.onChange).toHaveBeenCalledWith(234)
  })

  it('handles conversion for floats',()=>{
    render(<CustomNumberField schema={integerSchema} {...mockParams} />)

    fireEvent.change(screen.getByTestId('input-mock'),{target: {value: '23.4'}})
    expect(mockParams.onChange).toHaveBeenCalled()
    expect(mockParams.onChange).toHaveBeenCalledWith(23.4)
  })

  it('handles integer becoming bigint',()=>{
    const bigIntStr = `${Number.MAX_SAFE_INTEGER}123`
    const bigIntVal = BigInt(bigIntStr)

    render(<CustomNumberField schema={integerSchema} {...mockParams} formData={Number.MAX_SAFE_INTEGER} />)
    fireEvent.change(screen.getByTestId('input-mock'),{target: {value: bigIntStr}})

    expect(mockParams.onChange).toHaveBeenCalled()
    expect(mockParams.onChange).toHaveBeenCalledWith(bigIntVal)


  })


})
