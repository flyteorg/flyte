import { ClearAllFiltersButton } from '@/components/ClearAllFiltersButton'
import { useDateRangeParams } from '@/components/DatePicker'
import { Switch } from '@/components/Switch'
import { ProjectIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'
import { Filter, Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useOwnerFilter } from '@/hooks/filters/useOwnerFilter'
import { useOrg } from '@/hooks/useOrg'
import { useQueryFilters } from '@/hooks/useQueryFilters'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { getFilter } from '@/lib/filterUtils'
import { create } from '@bufbuild/protobuf'
import { ColumnDef } from '@tanstack/react-table'
import clsx from 'clsx'
import { subDays } from 'date-fns'
import { useParams } from 'next/navigation'
import { parseAsBoolean, useQueryState } from 'nuqs'
import { useCallback, useEffect, useMemo, type ReactNode } from 'react'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { StatusFilter } from '../StatusFilter'
import { ListRunsTableView } from './ListRunsTableView'
import { RunsTableRow } from './table/types'

type ProjectDomainParams = {
  domain?: string
  project?: string
}

export interface ListRunsContentProps {
  /** Optional array of additional Filter objects to apply to the run list */
  additionalFilters?: Filter[]
  /** Array of column definitions for the classic list view (required) */
  // 'any' because of complex type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  listTableColumns: ColumnDef<RunsTableRow, any>[]
  /** Optional CSS class name to apply to the container */
  className?: string
  /** Optional message to display when there are no runs to show */
  noRowsMessage?: string
  /** Optional flag to show/hide the "View by Group" toggle switch */
  showGroupToggle?: boolean
  /** When true, the last table row has no bottom border */
  hideLastRowBorder?: boolean
  /**
   * Optional element rendered directly above the filter chips. Pass a
   * fully-configured <DatePickerPopover> (or any node) — the caller controls
   * all its own props. Rendered inside the filter bar div so the parent's
   * [&>*:first-child] selectors always target a single consistent element.
   */
  datePickerPopover?: ReactNode
}

/**
 * Component that renders a list of runs with filtering and view options.
 * Supports both table and grouped views, with configurable filters and columns.
 *
 * @param props - The component props
 * @returns JSX element containing the runs list with filters and view options
 */
export const ListRunsContent = ({
  additionalFilters = [],
  listTableColumns,
  className,
  noRowsMessage = 'Get started by triggering a run with flyte from the CLI',
  showGroupToggle = false,
  hideLastRowBorder = false,
  datePickerPopover,
}: ListRunsContentProps) => {
  // Read route params and build projectId
  const params = useParams<ProjectDomainParams>()
  const organization = useOrg()

  // Read query params for filters
  const { filters: statusFilters, clearFilter: clearStatusFilter } =
    useQueryFilters()
  const { searchTerm, setSearchTerm } = useSearchTerm()
  const { selectedOwners, clearFilter: clearOwnerFilter } =
    useOwnerFilter('owner')
  const { dateRange, setDateRange, toParam } = useDateRangeParams()

  // Read query param for group view toggle
  // When showGroupToggle is true, use query param; task/trigger details default to table view
  const [viewByGroupQuery, setViewByGroupQuery] = useQueryState(
    'group',
    parseAsBoolean.withDefault(false),
  )

  // Force table view when showGroupToggle is false (task/trigger details pages)
  const viewByGroup = showGroupToggle ? viewByGroupQuery : false

  // Sync the DatePicker to "Last 7 days" when no existing date params provided
  useEffect(() => {
    if (!dateRange) {
      const now = new Date()
      setDateRange({ from: subDays(now, 7), to: now }, 'Last 7 days')
    }
  }, [dateRange, setDateRange])

  const projectId = useMemo(
    () =>
      params.domain && params.project && organization
        ? create(ProjectIdentifierSchema, {
            domain: params.domain,
            name: params.project,
            organization,
          })
        : undefined,
    [params.domain, params.project, organization],
  )

  // Build filters from query params
  const queryParamFilters = useMemo(() => {
    const list: Filter[] = []

    // Search term filter
    if (searchTerm) {
      list.push(
        getFilter({
          function: Filter_Function.CONTAINS_CASE_INSENSITIVE,
          field: 'task_name',
          values: [searchTerm],
        }),
      )
    }

    // Status filter
    if (statusFilters.status && statusFilters.status.length > 0) {
      const statusValues = statusFilters.status.map((s) =>
        ActionPhase[s as keyof typeof ActionPhase].toString(),
      )
      list.push(
        getFilter({
          function:
            statusValues.length > 1
              ? Filter_Function.VALUE_IN
              : Filter_Function.EQUAL,
          field: 'phase',
          values: statusValues,
        }),
      )
    }

    // Owner filter
    if (selectedOwners.length > 0) {
      list.push(
        getFilter({
          function: Filter_Function.VALUE_IN,
          field: 'created_by',
          values: selectedOwners,
        }),
      )
    }

    // Date range filters (created_at >= from AND created_at <= to)
    // Only add for table view - grouped view handles this internally with 30-day constraint
    if (!viewByGroup) {
      if (dateRange?.from) {
        list.push(
          getFilter({
            function: Filter_Function.GREATER_THAN_OR_EQUAL,
            field: 'created_at',
            values: [dateRange.from.toISOString()],
          }),
        )
      }
      // Only add "to" filter if it's NOT "now" - when toParam is "now", we want
      // live updates to show new runs, so we shouldn't cap the created_at filter
      if (dateRange?.to && toParam !== 'now') {
        list.push(
          getFilter({
            function: Filter_Function.LESS_THAN_OR_EQUAL,
            field: 'created_at',
            values: [dateRange.to.toISOString()],
          }),
        )
      }
    }

    return list
  }, [
    searchTerm,
    statusFilters.status,
    selectedOwners,
    dateRange?.from,
    dateRange?.to,
    toParam,
    viewByGroup,
  ])

  // Check if any filters are active
  const hasActiveFilters = useMemo(() => {
    return (
      (statusFilters.status && statusFilters.status.length > 0) ||
      !!searchTerm ||
      selectedOwners.length > 0
    )
  }, [statusFilters.status, searchTerm, selectedOwners.length])

  // Clear all filters
  const clearAllFilters = useCallback(() => {
    clearStatusFilter()
    setSearchTerm(null)
    clearOwnerFilter()
  }, [clearStatusFilter, setSearchTerm, clearOwnerFilter])

  // Merge query param filters with additional filters
  const allFilters = useMemo(
    () => [...queryParamFilters, ...additionalFilters],
    [queryParamFilters, additionalFilters],
  )

  // Create a key that changes when filters, projectId, or view type change to force remount
  // This ensures the component unmounts/remounts when dependencies change
  // Using a hash of the dependencies prevents unnecessary remounts when references change but values don't
  const viewKey = useMemo(() => {
    const projectKey = projectId
      ? `${projectId.organization}-${projectId.domain}-${projectId.name}`
      : 'no-project'
    // Stringify filters to create a stable key that only changes when filter values change
    const filtersKey = JSON.stringify(
      allFilters.map((f) => ({
        function: f.function,
        field: f.field,
        values: f.values,
      })),
    )
    return `${viewByGroup ? 'grouped' : 'table'}-${projectKey}-${filtersKey}`
  }, [projectId, allFilters, viewByGroup])

  return (
    <>
      <div
        className={clsx('flex h-full min-h-0 flex-grow flex-col', className)}
      >
        <div className="flex items-center justify-between gap-2 px-10 pb-6">
          <div className="flex flex-col gap-1">
            {datePickerPopover}
            <div className="flex items-center gap-2">
              <StatusFilter />

              {/* Clear all filters button */}
              {hasActiveFilters && (
                <ClearAllFiltersButton onClick={clearAllFilters} />
              )}
            </div>
          </div>

          {showGroupToggle && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-(--system-gray-5)">
                View by Group
              </span>
              <Switch
                data-testid="group-runs-toggle"
                size="sm"
                color="green"
                checked={viewByGroup}
                onChange={(checked) => setViewByGroupQuery(checked)}
              />
            </div>
          )}
        </div>
        {viewByGroup ? (
          <div className="flex min-w-0 flex-1 flex-col px-10">
            <LicensedEditionPlaceholder fullWidth title="Runs" />
          </div>
        ) : (
          <ListRunsTableView
            key={viewKey}
            projectId={projectId}
            filters={allFilters}
            columns={listTableColumns}
            noRowsMessage={noRowsMessage}
            enabled={true}
            enableLiveUpdates={toParam === 'now' || !dateRange}
            hideLastRowBorder={hideLastRowBorder}
          />
        )}
      </div>
    </>
  )
}
