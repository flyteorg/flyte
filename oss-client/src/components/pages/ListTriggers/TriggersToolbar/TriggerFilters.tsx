import { PopoverMenu } from '@/components/Popovers'

const mockFilterProps = {
  displayedValues: <></>,
  valuesCount: 0,
  onClearClick: () => {},
}

export const TriggerFilters = () => (
  <div className="flex flex-wrap gap-2">
    <PopoverMenu
      items={[{ id: 'trigger-status', label: 'Not implemented' }]}
      label="Status"
      variant="filter"
      filterProps={mockFilterProps}
    />
    <PopoverMenu
      items={[{ id: 'trigger-type', label: 'Not implemented' }]}
      label="Trigger type"
      variant="filter"
      filterProps={mockFilterProps}
    />
    <PopoverMenu
      items={[{ id: 'trigger-last-run', label: 'Not implemented' }]}
      label="Last run"
      variant="filter"
      filterProps={mockFilterProps}
    />
    <PopoverMenu
      items={[{ id: 'trigger-last-updated', label: 'Not implemented' }]}
      label="Last updated"
      variant="filter"
      filterProps={mockFilterProps}
    />
    <PopoverMenu
      items={[{ id: 'trigger-updated-by', label: 'Not implemented' }]}
      label="Updated by"
      variant="filter"
      filterProps={mockFilterProps}
    />
    <PopoverMenu
      items={[{ id: 'trigger-owner', label: 'Not implemented' }]}
      label="Owner"
      variant="filter"
      filterProps={mockFilterProps}
    />
  </div>
)
