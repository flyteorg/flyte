import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { vi } from 'vitest'
import { SetupBadgeFilter } from './StatusBadgeFilter'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

const mockToggleFilter = vi.fn()
const mockAdd = vi.fn()
const mockRemove = vi.fn()

vi.mock('@/hooks/useQueryFilters', () => ({
  useQueryFilters: () => ({
    addStatusFilters: mockAdd,
    removeStatusFilters: mockRemove,
    toggleFilter: mockToggleFilter,
    filters: { status: [] },
  }),
}))

vi.mock('@/components/StatusIcons', () => ({
  StatusIcon: ({ phase }: { phase: ActionPhase }) => (
    <div data-testid={`icon-${phase}`} />
  ),
}))

describe('SetupBadgeFilter', () => {
  const phaseCounts: Record<ActionPhase, number> = {
    [ActionPhase.UNSPECIFIED]: 0,
    [ActionPhase.QUEUED]: 2,
    [ActionPhase.INITIALIZING]: 3,
    [ActionPhase.WAITING_FOR_RESOURCES]: 1,
    [ActionPhase.RUNNING]: 0,
    [ActionPhase.SUCCEEDED]: 0,
    [ActionPhase.FAILED]: 0,
    [ActionPhase.ABORTED]: 0,
    [ActionPhase.TIMED_OUT]: 0,
  }

  it('renders the badge with total setup count', () => {
    render(<SetupBadgeFilter phaseCounts={phaseCounts} />)
    expect(screen.getByText('6')).toBeInTheDocument()
  })

  it('toggles all setup filters when clicked', () => {
    render(<SetupBadgeFilter phaseCounts={phaseCounts} />)
    fireEvent.click(screen.getByRole('button'))
    expect(mockAdd).toHaveBeenCalledWith([
      'QUEUED',
      'INITIALIZING',
      'WAITING_FOR_RESOURCES',
    ])
  })
})