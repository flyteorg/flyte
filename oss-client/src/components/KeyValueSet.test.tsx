import React from 'react'
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { KeyValueForm } from './KeyValueSet'

describe('KeyValueForm', () => {
  it('renders one empty row and Add button when state is empty', () => {
    const updateState = vi.fn()
    render(<KeyValueForm state={[]} updateState={updateState} />)
    expect(screen.getByPlaceholderText('Key')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Value')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /add/i })).toBeInTheDocument()
  })

  it('shows "Key is required" under the row when value is set but key is empty', () => {
    const updateState = vi.fn()
    render(
      <KeyValueForm
        state={[{ key: '', value: 'some-value' }]}
        updateState={updateState}
      />,
    )
    expect(screen.getByText('Key is required')).toBeInTheDocument()
  })

  it('shows "Key cannot contain spaces" when key has spaces', () => {
    const updateState = vi.fn()
    render(
      <KeyValueForm
        state={[{ key: 'key with spaces', value: 'v' }]}
        updateState={updateState}
      />,
    )
    expect(screen.getByText('Key cannot contain spaces')).toBeInTheDocument()
  })

  it('shows duplicate key error when two rows have the same key', () => {
    const updateState = vi.fn()
    render(
      <KeyValueForm
        state={[
          { key: 'same', value: '1' },
          { key: 'same', value: '2' },
        ]}
        updateState={updateState}
      />,
    )
    expect(screen.getAllByText(/Duplicate key/).length).toBeGreaterThanOrEqual(
      1,
    )
  })

  it('renders multiple rows when state has multiple pairs', () => {
    const updateState = vi.fn()
    render(
      <KeyValueForm
        state={[
          { key: 'a', value: '1' },
          { key: 'b', value: '2' },
        ]}
        updateState={updateState}
      />,
    )
    const keyInputs = screen.getAllByPlaceholderText('Key')
    const valueInputs = screen.getAllByPlaceholderText('Value')
    expect(keyInputs).toHaveLength(2)
    expect(valueInputs).toHaveLength(2)
    expect(keyInputs[0]).toHaveValue('a')
    expect(valueInputs[0]).toHaveValue('1')
    expect(keyInputs[1]).toHaveValue('b')
    expect(valueInputs[1]).toHaveValue('2')
  })
})
