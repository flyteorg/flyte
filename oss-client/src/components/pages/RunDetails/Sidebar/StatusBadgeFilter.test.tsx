import { vi } from 'vitest'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, beforeEach } from 'vitest'

const { mockToggle, mockClear, mockAdd, filtersStateRef } = vi.hoisted(() => {
  return {
    mockToggle: vi.fn(),
    mockClear: vi.fn(),
    mockAdd: vi.fn(),
    filtersStateRef: { status: [] as string[] },
  }
})

vi.mock('@/hooks/useQueryFilters', () => ({
  useQueryFilters: () => ({
    addStatusFilters: mockAdd,
    clearFilter: mockClear,
    toggleFilter: mockToggle,
    get filters() {
      return filtersStateRef
    },
  }),
}))

vi.mock('@/hooks/usePalette', () => ({
  useAccent: () => 'purple',
}))
vi.mock('@/lib/getColorByPhase', () => ({
  getColorsByPhase: () => 'purple',
}))
vi.mock('@/lib/numberUtils', () => ({
  formatNumber: (n: number) => n.toString(),
}))
vi.mock('@/components/StatusIcons', () => ({
  StatusIcon: ({ phase }: { phase: ActionPhase }) => (
    <div data-testid="status-icon">{phase}</div>
  ),
}))
vi.mock('@/components/StatusFilter', () => ({
  filterConfigs: [{ phase: ActionPhase.SUCCEEDED, value: 'SUCCEEDED' }],
}))
vi.mock('@/lib/phaseUtils', () => ({
  getPhaseFroπmEnum: () => 'SUCCEEDED',
}))

import { StatusBadgeFilter } from './StatusBadgeFilter'

describe('StatusBadgeFilter', () => {
  beforeEach(() => {
    filtersStateRef.status = []
    vi.clearAllMocks()
  })

  it('renders when phaseCount > 0', () => {
    render(
      <StatusBadgeFilter
        phaseCount={3}
        phase={ActionPhase.RUNNING}
        displayText="Running"
      />,
    )
    expect(screen.getByText('Running')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
  })

  it('renders nothing when phaseCount=0 and not visible', () => {
    const { container } = render(
      <StatusBadgeFilter phaseCount={0} phase={ActionPhase.ABORTED} />,
    )
    expect(container.firstChild).toBeNull()
  })

  it('forces visibility when forceVisible is true', () => {
    render(
      <StatusBadgeFilter
        phaseCount={0}
        phase={ActionPhase.ABORTED}
        forceVisible
        displayText="Aborted"
      />,
    )
    expect(screen.getByText('Aborted')).toBeInTheDocument()
  })

  it('calls handleClickCb when clicked', () => {
    const mockClick = vi.fn()
    render(
      <StatusBadgeFilter
        phaseCount={5}
        phase={ActionPhase.SUCCEEDED}
        displayText="Succeeded"
        handleClickCb={mockClick}
      />,
    )
    fireEvent.click(screen.getByRole('button'))
    expect(mockClick).toHaveBeenCalled()
  })

  it('renders active state when filter is included', () => {
    filtersStateRef.status = ['SUCCEEDED']
    render(
      <StatusBadgeFilter
        phaseCount={5}
        phase={ActionPhase.SUCCEEDED}
        displayText="Succeeded"
      />,
    )
    const button = screen.getByRole('button')
    expect(button).not.toHaveClass('plain')
  })

  it('clears and adds new filter when other filters exist', () => {
    filtersStateRef.status = ['FAILED']
    render(
      <StatusBadgeFilter
        phaseCount={4}
        phase={ActionPhase.SUCCEEDED}
        displayText="Succeeded"
      />,
    )
    fireEvent.click(screen.getByRole('button'))
    expect(mockClear).toHaveBeenCalled()
    expect(mockAdd).toHaveBeenCalledWith(['SUCCEEDED'])
  })
})