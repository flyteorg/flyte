import { ErrorBanner } from '@/components/ErrorBanner'
import { IORenderer } from '@/components/IORenderer/IORenderer'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { getFormDataFromSchemaDefaults } from '@/components/SchemaForm/utils'
import { TabSection } from '@/components/TabSection'
import ErrorPage from '@/components/pages/ErrorPage'
import { ActionDashboardLinks } from '@/components/pages/RunDetails/ActionDashboardLinks'
import { useSelectedActionId } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { useActionData } from '@/hooks/useActionData'
import { useLiteralToJson } from '@/hooks/useLiteralToJsonSchema'
import { useOrg } from '@/hooks/useOrg'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { getActionSpecInterface } from '@/lib/actionSpecUtils'
import {
  getTaskType,
  isActionSucceeded,
  isActionTerminal,
} from '@/lib/actionUtils'
import { isAbortInfoResult, isErrorInfoResult } from '@/lib/runTypeUtils'
import type { JSONSchema7 } from 'json-schema'
import { AnimatePresence, motion } from 'motion/react'
import { useParams } from 'next/navigation'
import { useCallback, useEffect, useMemo } from 'react'
import { stringify } from 'safe-stable-stringify'
import { ActionFanoutTable } from './ActionFanoutTable'
import { ActionMetadata } from './ActionMetadata'
import { ActionTimeline } from './ActionPhaseTimeline'
import { useGetFanoutData } from './hooks/useGetFanoutData'
import { useSelectedAttemptStore } from './state/AttemptStore'
import { useRunStore } from './state/RunStore'
import { RunDetailsPageParams } from './types'

interface UseActionDataWithJsonReturn {
  inputsJsonSchema: ReturnType<typeof useLiteralToJson>['data'] | undefined
  outputsJsonSchema: ReturnType<typeof useLiteralToJson>['data'] | undefined
  inputRawValues: Record<string, unknown> | undefined
  outputRawValues: Record<string, unknown> | undefined
  isInputsLoading: boolean
  isOutputsLoading: boolean
}

const useActionDataWithJson = ({
  actionDetails,
}: {
  actionDetails: ActionDetails | undefined
}): UseActionDataWithJsonReturn => {
  const { inputs, outputs } = getActionSpecInterface(actionDetails)

  const {
    data: actionData,
    isLoading: isActionDataLoading,
    isFetched: isActionDataFetched,
    refetch,
  } = useActionData({
    actionDetails,
    enabled: !!actionDetails,
  })

  const isTerminal = isActionTerminal(actionDetails)

  // Refetch when action succeeds to ensure outputs are available
  useEffect(() => {
    if (isActionSucceeded(actionDetails)) {
      refetch()
    }
  }, [actionDetails, refetch])

  const hasInputLiterals = !!actionData?.inputs?.literals?.length
  const hasOutputLiterals = !!actionData?.outputs?.literals?.length

  // Convert literals to JSON schema for inputs
  const inputsJsonSchema = useLiteralToJson(
    hasInputLiterals && inputs && actionData?.inputs
      ? {
          literals: actionData.inputs.literals,
          variables: inputs,
        }
      : null,
  )

  // Convert literals to JSON schema for outputs
  const outputsJsonSchema = useLiteralToJson(
    hasOutputLiterals && outputs && actionData?.outputs
      ? {
          literals: actionData.outputs.literals,
          variables: outputs,
        }
      : null,
  )

  // Extract raw values from the schemas
  const inputRawValues = useMemo(() => {
    const json = inputsJsonSchema.data?.json as JSONSchema7 | undefined
    const raw = json ? getFormDataFromSchemaDefaults(json) : undefined
    return raw && typeof raw === 'object' && !Array.isArray(raw)
      ? (raw as Record<string, unknown>)
      : undefined
  }, [inputsJsonSchema.data?.json])

  const outputRawValues = useMemo(() => {
    const json = outputsJsonSchema.data?.json as JSONSchema7 | undefined
    const raw = json ? getFormDataFromSchemaDefaults(json) : undefined
    return raw && typeof raw === 'object' && !Array.isArray(raw)
      ? (raw as Record<string, unknown>)
      : undefined
  }, [outputsJsonSchema.data?.json])

  // For inputs: if actionData is fetched and there are no input literals, show "No input data"
  // (inputs are always available from the start, unlike outputs)
  const shouldShowInputMessage = isActionDataFetched && !hasInputLiterals

  // For outputs: we should not show loading spinner in these cases:
  // 1. Action is terminal and actionData is fetched but outputs are empty/undefined (show "No output data")
  // 2. Action is not terminal and we're waiting for outputs (show "Waiting for output...")
  // In both cases, we want to show the message instead of a loading spinner
  const shouldShowOutputMessage =
    (!isTerminal && !hasOutputLiterals) ||
    (isTerminal && isActionDataFetched && !hasOutputLiterals)

  return {
    inputsJsonSchema: inputsJsonSchema.data,
    outputsJsonSchema: outputsJsonSchema.data,
    inputRawValues,
    outputRawValues,
    isInputsLoading: shouldShowInputMessage
      ? false // Don't show loading spinner, show message instead
      : isActionDataLoading ||
        inputsJsonSchema.isLoading ||
        !inputsJsonSchema.isFetched,
    isOutputsLoading: shouldShowOutputMessage
      ? false // Don't show loading spinner, show message instead
      : isActionDataLoading ||
        outputsJsonSchema.isLoading ||
        !outputsJsonSchema.isFetched,
  }
}

export interface RunDetailsSummaryTabProps {
  selectedActionDetailsQuery: ReturnType<typeof useWatchActionDetails>
}
export const RunDetailsSummaryTab = ({
  selectedActionDetailsQuery,
}: RunDetailsSummaryTabProps) => {
  const run = useRunStore((s) => s.run)
  const selectedActionId = useSelectedActionId()
  const selectedActionDetails = selectedActionDetailsQuery.data
  const selectedAttempt = useSelectedAttemptStore((s) => s.selectedAttempt)
  const params = useParams<RunDetailsPageParams>()
  const org = useOrg()
  const {
    inputsJsonSchema,
    outputsJsonSchema,
    inputRawValues,
    outputRawValues,
    isInputsLoading,
    isOutputsLoading,
  } = useActionDataWithJson({
    actionDetails: selectedActionDetails,
  })

  const fadeKey = `${selectedActionId}-${selectedAttempt?.attempt ?? 'none'}`
  const taskType = getTaskType(selectedActionDetails?.metadata)
  const actionId = selectedActionDetails?.id?.name
  const fanoutAction = useGetFanoutData()
  const shouldShowFanoutTable =
    fanoutAction && Object.values(fanoutAction).length > 0

  const isRunTerminal = isActionTerminal(run?.action)

  const runName =
    run?.action?.id?.run?.name ||
    selectedActionDetails?.id?.run?.name ||
    params?.runId

  /** Format inputs/outputs object for copy (JSON stringified). */
  const formatJsonForCopy = useCallback(
    (values: Record<string, unknown>): string => {
      try {
        const result = stringify(values, null, 2)
        return result ?? '{}'
      } catch {
        return '{}'
      }
    },
    [],
  )

  const outputsCopyContent = useMemo(() => {
    if (isOutputsLoading || !outputRawValues) return undefined
    return formatJsonForCopy(outputRawValues)
  }, [isOutputsLoading, outputRawValues, formatJsonForCopy])

  const inputsCopyContent = useMemo(() => {
    if (isInputsLoading || !inputRawValues) return undefined
    return formatJsonForCopy(inputRawValues)
  }, [isInputsLoading, inputRawValues, formatJsonForCopy])

  if (selectedActionDetailsQuery.error) {
    return <ErrorPage />
  }

  if (
    !selectedActionDetailsQuery.isFetched ||
    !selectedActionDetailsQuery.data
  ) {
    return <LoadingSpinner />
  }
  return (
    <div className="relative flex h-full min-h-0 w-full min-w-0">
      <AnimatePresence mode="wait">
        <motion.div
          key={fadeKey}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.3, ease: 'easeInOut' }}
          className={`absolute inset-0 flex flex-col gap-6 px-8 ${taskType === 'trace' ? 'pt-7' : 'pt-2.5'} [&>*:last-child]:pb-8`}
        >
          {actionId &&
          selectedActionDetails?.metadata?.spec?.case === 'task' ? (
            <ActionMetadata
              actionDetails={selectedActionDetails}
              cluster={selectedAttempt?.cluster}
            />
          ) : null}
          {selectedAttempt?.errorInfo ||
          isErrorInfoResult(selectedActionDetails?.result) ||
          isAbortInfoResult(selectedActionDetails?.result) ? (
            <div className="flex-shrink-0">
              <ErrorBanner
                attempt={selectedAttempt}
                result={selectedActionDetails?.result}
                org={org}
                domain={params?.domain}
                project={params?.project}
                runName={runName}
                actionId={actionId}
              />
            </div>
          ) : null}
          <ActionTimeline
            phaseTransitions={selectedAttempt?.phaseTransitions}
            taskType={taskType}
          />
          {shouldShowFanoutTable && (
            <TabSection heading="Sub actions">
              <ActionFanoutTable
                childPhaseCounts={fanoutAction}
                isRunTerminal={isRunTerminal}
              />
            </TabSection>
          )}
          <ActionDashboardLinks attempt={selectedAttempt} />
          <TabSection
            copyButtonContent={outputsCopyContent}
            heading="Output"
            defaultView="raw"
            sectionContent={({ isRawView }) => (
              <IORenderer
                jsonSchema={outputsJsonSchema?.json}
                formData={outputRawValues}
                formDataProvided
                isLoading={isOutputsLoading}
                isRawView={isRawView}
                noDataMessage={
                  isActionTerminal(selectedActionDetails)
                    ? 'No output data'
                    : 'Waiting for output...'
                }
              />
            )}
            showRawJsonToggle
          />
          <TabSection
            copyButtonContent={inputsCopyContent}
            heading="Input"
            defaultView="raw"
            sectionContent={({ isRawView }) => (
              <IORenderer
                jsonSchema={inputsJsonSchema?.json}
                formData={inputRawValues}
                formDataProvided
                isLoading={isInputsLoading}
                isRawView={isRawView}
                noDataMessage="No input data"
              />
            )}
            showRawJsonToggle
          />
        </motion.div>
      </AnimatePresence>
    </div>
  )
}
