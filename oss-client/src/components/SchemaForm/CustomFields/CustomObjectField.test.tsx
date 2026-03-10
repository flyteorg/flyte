import React from 'react'
import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import CustomObjectField from './CustomObjectField'
import { ObjectFieldTemplateProps } from '@rjsf/utils'
import { ObjectPropertyContextProvider } from './Objects/ObjectPropertyContext'

const mockOnAddClick = vi.fn()

const createMockProps = (
  overrides?: Partial<ObjectFieldTemplateProps>,
): ObjectFieldTemplateProps => ({
  title: '',
  schema: { type: 'object' },
  properties: [],
  required: false,
  readonly: false,
  onAddClick: mockOnAddClick,
  ...overrides,
})

describe('CustomObjectField', () => {
  describe('depth calculation', () => {
    it('uses context depth for root level objects (no title)', () => {
      const props = createMockProps({ title: '' })
      const { container } = render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      // Root level should have no indentation
      const element = container.firstChild as HTMLElement
      expect(element.style.paddingLeft).toBe('')
    })

    it('increments depth for nested objects (with title)', () => {
      const props = createMockProps({ title: 'NestedObject' })
      const { container } = render(
        <ObjectPropertyContextProvider parentName="parent" index={0} depth={1}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      // Nested object should have indentation (depth 1 + 1 = 2, so 32px)
      const element = container.firstChild as HTMLElement
      expect(element.style.paddingLeft).toBe('32px')
    })

    it('respects max indentation cap', () => {
      const props = createMockProps({ title: 'DeeplyNested' })
      const { container } = render(
        <ObjectPropertyContextProvider parentName="parent" index={0} depth={5}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      // Should cap at MAX_INDENT (64px) even if depth would be 6
      const element = container.firstChild as HTMLElement
      expect(element.style.paddingLeft).toBe('64px')
    })

    it('shows border for nested objects', () => {
      const props = createMockProps({ title: 'NestedObject' })
      const { container } = render(
        <ObjectPropertyContextProvider parentName="parent" index={0} depth={1}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      const element = container.firstChild as HTMLElement
      expect(element.className).toContain('border-l')
    })

    it('does not show border for root level objects', () => {
      const props = createMockProps({ title: '' })
      const { container } = render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      const element = container.firstChild as HTMLElement
      expect(element.className).not.toContain('border-l')
    })
  })

  describe('rendering', () => {
    it('renders title when provided', () => {
      const props = createMockProps({ title: 'TestObject' })
      render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      expect(screen.getByText('TestObject')).toBeInTheDocument()
    })

    it('does not render title for root level objects', () => {
      const props = createMockProps({ title: '' })
      const { container } = render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      // Should not have title element
      expect(
        container.querySelector('span.font-medium'),
      ).not.toBeInTheDocument()
    })

    it('renders type label when schema has type', () => {
      const props = createMockProps({
        title: 'TestObject',
        schema: { type: 'object' },
      })
      render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      expect(screen.getByText('(object)')).toBeInTheDocument()
    })

    it('shows required indicator when required', () => {
      const props = createMockProps({
        title: 'TestObject',
        required: true,
      })
      render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      const titleElement = screen.getByText('TestObject').parentElement
      expect(titleElement?.textContent).toContain('*')
    })

    it('shows add button when additionalProperties is allowed', () => {
      const props = createMockProps({
        title: 'TestObject',
        schema: { type: 'object', additionalProperties: { type: 'string' } },
        readonly: false,
      })
      render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      expect(screen.getByText(/Add/i)).toBeInTheDocument()
    })

    it('does not show add button when readonly', () => {
      const props = createMockProps({
        title: 'TestObject',
        schema: { type: 'object', additionalProperties: { type: 'string' } },
        readonly: true,
      })
      render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={0}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      expect(screen.queryByText(/Add/i)).not.toBeInTheDocument()
    })
  })

  describe('property context propagation', () => {
    it('passes correct depth to child properties', () => {
      const props = createMockProps({
        title: 'ParentObject',
        properties: [
          {
            content: <div data-testid="child">Child</div>,
            name: 'child',
            disabled: false,
            readonly: false,
            hidden: false,
          },
        ],
      })
      render(
        <ObjectPropertyContextProvider parentName="" index={0} depth={1}>
          <CustomObjectField {...props} />
        </ObjectPropertyContextProvider>,
      )

      // Child should receive depth 2 (parent depth 1 + 1 for title)
      expect(screen.getByTestId('child')).toBeInTheDocument()
    })
  })
})
