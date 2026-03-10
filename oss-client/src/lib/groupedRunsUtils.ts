import { RunsTableRow } from '@/components/ListRuns/table/types'
import { TaskGroup } from '@/gen/flyteidl2/workflow/run_definition_pb'

export type FlattenedRow =
  // Represents either a group row
  | {
      type: 'group'
      data: TaskGroup
      groupKey: string
      id: string
    }
  // Represents the header for runs within a group
  | {
      type: 'runsHeader'
      groupName: string
      id: string
    }
  // Represents an individual run within a group
  | {
      type: 'run'
      data: RunsTableRow
      groupName: string
      id: string
    }
  // Represents a "Load More" button for loading additional runs within a group
  | {
      type: 'loadMore'
      groupName: string
      id: string
      hasMore: boolean
      isLoading: boolean
      onLoadMore: () => void
    }

/**
 * Calculates the start time for watch queries (default 7 days with 5min buffer).
 * Used to ensure useWatchGroups and useWatchRuns use the exact same time window.
 *
 * @param days Number of days to look back (default: 7)
 * @param referenceTime Optional reference time. If not provided, uses current time.
 * @returns A Date object representing the start time (days * 24 hours ago + 5 minute buffer)
 */
export function getStartTime(
  days: number = 7,
  referenceTime: Date = new Date(),
): Date {
  const hoursInDays = days * 24
  const startDateMs = referenceTime.getTime() - hoursInDays * 60 * 60 * 1000 // days in ms
  return new Date(startDateMs)
}

/**
 * Generate a unique key for a group using environmentName and taskName
 * Format: environmentName.taskName
 */
export function getGroupKey(group: TaskGroup): string {
  // Use environmentName in group keys to avoid collisions across environments
  return `${group.environmentName}.${group.taskName}`
}

/**
 * Transforms group runs data by removing the query property.
 * This is used to convert the internal data structure to the format expected by flattenGroupRows.
 */
export function transformGroupRunsData(
  groupRunsData: Map<
    string,
    {
      runs: RunsTableRow[]
      hasMore: boolean
      isLoading: boolean
      query?: unknown
    }
  >,
): Map<
  string,
  {
    runs: RunsTableRow[]
    hasMore: boolean
    isLoading: boolean
  }
> {
  return new Map(
    Array.from(groupRunsData.entries()).map(([key, value]) => [
      key,
      {
        runs: value.runs,
        hasMore: value.hasMore,
        isLoading: value.isLoading,
      },
    ]),
  )
}

/**
 * Flattens groups and their runs into a single array for the unified table.
 * Groups are always included, and if a group is expanded, its runs are included as well.
 * Only one group can be expanded at a time.
 *
 * @param groups Array of task groups to flatten
 * @param expandedGroupKey The key of the currently expanded group (if any)
 * @param groupRunsData Map of group keys to their runs data
 * @param handleLoadMore Callback function to load more runs for a group
 * @returns Array of flattened rows (groups, runs headers, runs, and load more buttons)
 */
export function flattenGroupRows(
  groups: TaskGroup[],
  expandedGroupKey: string | null,
  groupRunsData: Map<
    string,
    {
      runs: RunsTableRow[]
      hasMore: boolean
      isLoading: boolean
    }
  >,
  handleLoadMore: (groupKey: string) => void,
): FlattenedRow[] {
  const rows: FlattenedRow[] = []

  groups.forEach((group) => {
    // Store group key for later use
    const groupKey = getGroupKey(group)
    rows.push({
      type: 'group',
      data: group,
      groupKey,
      id: `group-${groupKey}`,
    })

    // If expanded, add runs header, run rows, and load more button
    if (expandedGroupKey === groupKey) {
      const runsData = groupRunsData.get(groupKey)
      if (runsData?.runs.length) {
        rows.push({
          type: 'runsHeader',
          groupName: groupKey,
          id: `runsHeader-${groupKey}`,
        })

        runsData.runs.forEach((run, index) => {
          rows.push({
            type: 'run',
            data: run,
            groupName: groupKey,
            id: `run-${groupKey}-${index}`,
          })
        })

        if (runsData.hasMore) {
          rows.push({
            type: 'loadMore',
            groupName: groupKey,
            id: `loadMore-${groupKey}`,
            hasMore: runsData.hasMore,
            isLoading: runsData.isLoading,
            onLoadMore: () => handleLoadMore(groupKey),
          })
        }
      }
    }
  })

  return rows
}
