import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import {
  DescriptionListWrapper,
  type Section,
  type SectionItem,
} from './DescriptionListWrapper'

describe('DescriptionListWrapper', () => {
  it('renders section items with name and value in pretty view', () => {
    const sections: Section[] = [
      {
        id: 'main',
        name: 'Main',
        items: [
          { name: 'Name', value: 'my-task' },
          { name: 'Version', value: 'v1', copyBtn: true },
        ],
      },
    ]
    render(
      <DescriptionListWrapper
        isRawView={false}
        sections={sections}
        rawJson={{}}
      />,
    )
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('my-task')).toBeInTheDocument()
    expect(screen.getByText('Version')).toBeInTheDocument()
    expect(screen.getByText('v1')).toBeInTheDocument()
  })

  it('renders item with url as a link with expected href and value text', () => {
    const sections: Section[] = [
      {
        id: 'main',
        name: '',
        items: [
          {
            name: 'Image',
            value: 'my-registry.io/image:tag',
            url: '/v2/domain/d/project/p/runs/run-123',
            copyBtn: true,
          },
        ],
      },
    ]
    render(
      <DescriptionListWrapper
        isRawView={false}
        sections={sections}
        rawJson={{}}
      />,
    )
    const link = screen.getByRole('link', { name: 'my-registry.io/image:tag' })
    expect(link).toBeInTheDocument()
    expect(link).toHaveAttribute('href', '/v2/domain/d/project/p/runs/run-123')
    expect(link).toHaveTextContent('my-registry.io/image:tag')
  })

  it('when item has url and copyBtn, renders both link and copy button with value for copy', () => {
    const sections: Section[] = [
      {
        id: 'main',
        name: '',
        items: [
          {
            name: 'Image',
            value: 'foo/bar:latest',
            url: '/runs/xyz',
            copyBtn: true,
          },
        ],
      },
    ]
    render(
      <DescriptionListWrapper
        isRawView={false}
        sections={sections}
        rawJson={{}}
      />,
    )
    const link = screen.getByRole('link', { name: 'foo/bar:latest' })
    expect(link).toHaveAttribute('href', '/runs/xyz')
    const copyButton = screen.getByTitle('Copy to clipboard')
    expect(copyButton).toBeInTheDocument()
  })

  it('renders item without url as plain text (no anchor)', () => {
    const sections: Section[] = [
      {
        id: 'main',
        name: '',
        items: [{ name: 'Plain', value: 'just text', copyBtn: false }],
      },
    ]
    render(
      <DescriptionListWrapper
        isRawView={false}
        sections={sections}
        rawJson={{}}
      />,
    )
    expect(screen.getByText('just text')).toBeInTheDocument()
    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })

  it('renders item with url and undefined value as link with empty content', () => {
    const sections: Section[] = [
      {
        id: 'main',
        name: '',
        items: [
          {
            name: 'Image',
            value: undefined,
            url: '/v2/domain/d/project/p/runs/run-456',
            copyBtn: false,
          } as SectionItem,
        ],
      },
    ]
    render(
      <DescriptionListWrapper
        isRawView={false}
        sections={sections}
        rawJson={{}}
      />,
    )
    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/v2/domain/d/project/p/runs/run-456')
    // In the default (non-fullWidth) layout, undefined values render as \"-\" via renderUnknownValue
    expect(link).toHaveTextContent('-')
  })

  it('renders url item in fullWidthItemBorders section with correct link', () => {
    const sections: Section[] = [
      {
        id: 'custom',
        name: 'Custom',
        fullWidthItemBorders: true,
        items: [
          {
            name: 'Blob',
            value: 's3://bucket/key',
            url: 'https://console.aws.amazon.com/s3/buckets/bucket',
            copyBtn: true,
          },
        ],
      },
    ]
    render(
      <DescriptionListWrapper
        isRawView={false}
        sections={sections}
        rawJson={{}}
      />,
    )
    const link = screen.getByRole('link', { name: 's3://bucket/key' })
    expect(link).toHaveAttribute(
      'href',
      'https://console.aws.amazon.com/s3/buckets/bucket',
    )
  })

  it('renders raw JSON view when isRawView is true', () => {
    const rawJson = { task: 'spec' }
    render(
      <DescriptionListWrapper
        isRawView={true}
        sections={[]}
        rawJson={rawJson}
      />,
    )
    expect(screen.getByTestId('dl-wrapper')).toBeInTheDocument()
    // Sections (pretty) content should not show section labels
    expect(screen.queryByText('Name')).not.toBeInTheDocument()
  })
})
