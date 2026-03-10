import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  ActionAttempt,
  ActionDetails,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { beforeEach, describe, it, expect, vi } from 'vitest'

import { useSelectedActionId } from './hooks/useSelectedItem'
import {
  SelectedAttemptState,
  useSelectedAttemptStore,
} from './state/AttemptStore'
import {
  type PopoverMenuProps,
  MenuItem,
} from '@/components/Popovers/PopoverMenu'

vi.mock('./hooks/useSelectedItem', () => ({
  useSelectedActionId: vi.fn(),
}))
vi.mock('./state/AttemptStore', () => ({
  useSelectedAttemptStore: vi.fn(),
}))

vi.mock('@/components/Popovers', () => ({
  PopoverMenu: ({ items, children, trigger }: PopoverMenuProps) => (
    <div data-testid="popover">
      {trigger}
      {items?.map((item: MenuItem) => (
        <button
          key={item.id}
          onClick={item.onClick}
          data-testid={`menu-item-${item.id}`}
        >
          {item.label}
        </button>
      ))}
      {children}
    </div>
  ),
}))

import { AttemptPicker } from './AttemptPicker'

const actionDetailsMock = {
  attempts: [
    { attempt: 1 } as ActionAttempt,
    { attempt: 2 } as ActionAttempt,
    { attempt: 3 } as ActionAttempt,
  ],
} as ActionDetails

vi.mock('@/components/icons/ArrowsRightIcon', () => ({
  ArrowsRightIcon: () => <span data-testid="arrows-icon" />,
}))

describe('AttemptPicker', () => {
  const mockSetSelectedAttempt = vi.fn()

  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(useSelectedActionId).mockReturnValue('action-123')
    vi.mocked(useSelectedAttemptStore).mockImplementation((selector) => {
      const store: SelectedAttemptState = {
        selectedAttemptNumber: 1,
        selectedAttempt: { attempt: 1 } as ActionAttempt,
        setSelectedAttempt: mockSetSelectedAttempt,
        clearSelectedAttempt: vi.fn(),
        manuallySelectedAttemptNumber: 'latest',
        clearManuallySelectedAttempt: vi.fn(),
      }

      return typeof selector === 'function' ? selector(store) : store
    })
  })

  it('returns null when there are no attempts', () => {
    const { container } = render(
      <AttemptPicker
        actionDetails={{ attempts: [] } as unknown as ActionDetails}
      />,
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders current attempt info when attempts exist', () => {
    render(<AttemptPicker actionDetails={actionDetailsMock} />)

    expect(screen.getByTestId('attempt-picker-trigger')).toBeInTheDocument()
    expect(screen.getByText(/Attempt 1\/3/)).toBeVisible()
  })

  it('builds menu items for each attempt', async () => {
    render(<AttemptPicker actionDetails={actionDetailsMock} />)

    expect(screen.getByTestId('menu-item-Attempt-1')).toBeVisible()
  })

  it('calls setSelectedAttempt when clicking an attempt item', async () => {
    const user = userEvent.setup()
    render(<AttemptPicker actionDetails={actionDetailsMock} />)

    await user.click(screen.getByTestId('menu-item-Attempt-2'))

    expect(mockSetSelectedAttempt).toHaveBeenCalledWith({
      attempt: { attempt: 2 },
      isManual: true,
    })
  })
})
