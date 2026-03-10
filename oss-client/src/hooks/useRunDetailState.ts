import { useRunStore } from '@/components/pages/RunDetails/state/RunStore'
import { Filter, Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { EnrichedAction } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { RunService } from '@/gen/flyteidl2/workflow/run_service_pb'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { isActionTerminal } from '@/lib/actionUtils'
import { useQueryState } from 'nuqs'
import { useEffect } from 'react'
import { useConnectRpcClient } from './useConnectRpc'

const getFilters = (
  filterString: string | null,
  searchTerm: string | null,
): Filter[] => {
  const newFilters: Filter[] = []
  if (filterString) {
    const phases = filterString.split(';').filter((p) => !!p)
    const phaseFilter = {
      field: 'PHASE',
      function: Filter_Function.VALUE_IN,
      values: [] as string[],
    } as Filter

    phases.forEach((p) => {
      // Convert phase string to enum value, then back to the expected format
      const enumKey = p as keyof typeof ActionPhase
      if (ActionPhase[enumKey] !== undefined) {
        phaseFilter.values.push(ActionPhase[enumKey].toString())
      }
    })
    if (phaseFilter.values.length > 0) {
      newFilters.push(phaseFilter)
    }
  }
  if (searchTerm) {
    const searchFilter = {
      field: 'NAME',
      function: Filter_Function.CONTAINS_CASE_INSENSITIVE,
      values: [searchTerm],
    } as Filter
    newFilters.push(searchFilter)
  }
  return newFilters
}

export const useRunDetailState = ({
  domain,
  orgId,
  projectId,
  runId,
}: {
  domain: string
  orgId: string
  projectId: string
  runId: string
}) => {
  const client = useConnectRpcClient(RunService)
  const { searchTerm } = useSearchTerm()
  const [statusValue] = useQueryState('status')

  const upsert = useRunStore((s) => s.upsertAction)
  const reset = useRunStore((s) => s.reset)

  useEffect(() => {
    let controller = new AbortController()
    let lastEvent = Date.now()
    let buffer: EnrichedAction[] = []
    let flushInterval: NodeJS.Timeout | null = null

    const flushBuffer = () => {
      if (buffer.length > 0) {
        const batch = [...buffer.flat()]
        buffer = []
        upsert(batch)
      }
    }

    const start = () => {
      controller = new AbortController()
      flushInterval = setInterval(() => {
        flushBuffer()
      }, 250)
      ;(async () => {
        try {
          const stream = client.watchActions(
            {
              runId: {
                domain,
                name: runId,
                org: orgId,
                project: projectId,
              },
              ...((statusValue || searchTerm) && {
                filter: getFilters(statusValue, searchTerm),
              }),
            },
            { signal: controller.signal },
          )

          for await (const event of stream) {
            lastEvent = Date.now()
            buffer.push(...event.enrichedActions)
          }
        } catch (e) {
          if (!controller.signal.aborted) {
            console.error('error watching actions', e)
          }
        }
      })()
    }

    start()

    // Heartbeat: check every 5s, reset if >10s since last event
    const heartbeat = setInterval(() => {
      const isTerminal = isActionTerminal(useRunStore.getState()?.run?.action)
      // stop polling if we've reached terminal state
      if (isTerminal) {
        controller.abort()
        clearInterval(heartbeat)
        if (flushInterval) {
          clearInterval(flushInterval)
          flushInterval = null
        }
        flushBuffer()
        return
      }
      // after 5 minutes with no activity, restart the connection
      if (Date.now() - lastEvent > 5 * 60 * 1000) {
        console.warn('Stream stalled, restarting…')
        controller.abort()
        if (flushInterval) {
          clearInterval(flushInterval)
          flushInterval = null
        }
        flushBuffer()
        start()
      }
    }, 5000)

    return () => {
      clearInterval(heartbeat)
      if (flushInterval) {
        clearInterval(flushInterval)
        flushInterval = null
      }
      flushBuffer()
      controller.abort()
      reset()
    }
  }, [
    client,
    domain,
    runId,
    orgId,
    projectId,
    searchTerm,
    statusValue,
    upsert,
    reset,
  ])
}

export default useRunDetailState
