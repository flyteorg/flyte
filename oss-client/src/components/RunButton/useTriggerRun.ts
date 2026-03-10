import { TriggerName } from '@/gen/flyteidl2/common/identifier_pb'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'
import { useListTaskVersions } from '@/hooks/useListTaskVersions'
import { Sort, Sort_Direction } from '@/gen/flyteidl2/common/list_pb'
import { getSortParamForQueryKey } from '@/hooks/useQueryParamSort'
import { useTaskLaunchFormData } from '@/components/LaunchForm/hooks/useTaskLaunchFormData'

const defaultSort: Sort = {
  key: 'created_at',
  direction: Sort_Direction.DESCENDING,
} as Sort

export function useTriggerRun(triggerName: TriggerName | undefined) {
  const { buttonText, setIsOpen, setTriggerName, isOpen } = useLaunchFormState()

  // Fetch latest version of the associated task
  const allVersionsQuery = useListTaskVersions({
    taskName: triggerName?.taskName || '',
    project: triggerName?.project,
    domain: triggerName?.domain,
    sort: {
      sortBy: defaultSort,
      sortForQueryKey: getSortParamForQueryKey(defaultSort),
    },
  })
  const latestVersion =
    allVersionsQuery.data?.pages?.[0]?.versions?.[0]?.version
  const isVersionsFetched = allVersionsQuery.isFetched

  // Use task launch form data to populate the form with task inputs
  // Pass task parameters from triggerName so the hook works correctly on TriggerDetails page
  const { drawerMeta, isDataFetched, formMethods } = useTaskLaunchFormData({
    version: latestVersion || '',
    taskName: triggerName?.taskName,
    project: triggerName?.project,
    domain: triggerName?.domain,
  })

  const isReady =
    triggerName && latestVersion && isVersionsFetched && isDataFetched
  const isLoading =
    allVersionsQuery.isLoading ||
    !isVersionsFetched ||
    (!!latestVersion && !isDataFetched)

  return {
    buttonText,
    setTriggerName,
    triggerName,
    isOpen,
    setIsOpen,
    latestVersion,
    isVersionsFetched,
    isDataFetched,
    isReady,
    isLoading,
    drawerMeta,
    formMethods,
  }
}
