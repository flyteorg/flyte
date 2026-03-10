import { Badge } from '@/components/Badge'
import SchemaForm from '@/components/SchemaForm'
import { ToggleButtonGroup } from '@/components/ToggleButtonGroup'
import { useDefaultInputsJson } from '@/hooks/useDefaultInputsJson'
import { useOrg } from '@/hooks/useOrg'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { useTaskSpecLaunchForm } from '@/hooks/useTaskSpecLaunchForm'
import { ProjectDomainPageParams } from '@/types/pageParams'
import { registerFlyteLightTheme, FLYTE_LIGHT_THEME } from '@/utils/monacoThemes'
import MonacoEditor from '@monaco-editor/react'
import { RJSFValidationError } from '@rjsf/utils'
import { merge } from 'lodash'
import { useParams } from 'next/navigation'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useFormContext } from 'react-hook-form'
import stringify from 'safe-stable-stringify'
import { TaskDetails } from '../types'
import { CreateTriggerState, FormData } from './types'

type InputMode = 'raw' | 'pretty'

const initialRjsfFormState = { errors: [], touchedFields: {} }

type RjsfSchemaFormState = {
  errors: RJSFValidationError[]
  touchedFields: Record<string, boolean>
}

export const TriggerInputs = ({
  taskDetails,
}: {
  taskDetails: TaskDetails
}) => {
  const [inputMode, setInputMode] = useState<InputMode>('pretty')
  const { project, domain } = useParams<ProjectDomainPageParams>()
  const [rjsfFormState, setRjsfFormState] =
    useState<RjsfSchemaFormState>(initialRjsfFormState)
  const org = useOrg()
  const { setValue, trigger, watch, setError, clearErrors, formState } =
    useFormContext<CreateTriggerState>()
  const formDataValues = watch('formData')
  const kickoffTimeInputArg = watch('kickoffTimeInputArg')

  const setKickoffTimeInputArg = useCallback(
    (fieldName: string | undefined) => {
      setValue('kickoffTimeInputArg', fieldName)
    },
    [setValue],
  )

  const datetimeFieldConfig = useMemo(
    () => ({
      selectedField: kickoffTimeInputArg,
      setSelectedField: setKickoffTimeInputArg,
      checkboxLabel: 'Use run’s start time',
      disabledPlaceholder: 'Using run’s start time',
    }),
    [kickoffTimeInputArg, setKickoffTimeInputArg],
  )

  const taskDetailsQuery = useTaskDetails({
    name: taskDetails.taskId,
    version: taskDetails?.taskVersion,
    project,
    domain,
    org,
  })

  const taskQuery = useTaskSpecLaunchForm({
    taskSpec: taskDetailsQuery.data?.details?.spec,
    enabled: !!taskDetailsQuery.data?.details?.spec,
  })

  const literalsQuery = useDefaultInputsJson(
    taskDetails.taskVersion,
    taskDetails.taskId,
    project,
    domain,
  )
  const jsonInputs = merge(
    {},
    taskQuery.data?.json ?? {},
    literalsQuery.data?.json ?? {},
  )

  useEffect(() => {
    if (jsonInputs) {
      setValue('inputs', jsonInputs)
      clearErrors('inputs')
    }
  }, [clearErrors, jsonInputs, setValue])

  const onRjsfChange = useCallback(
    (data: FormData) => {
      setValue('formData', data)
      trigger('formData')
    },
    [setValue, trigger],
  )

  return (
    <div>
      <div className="flex justify-between py-2">
        <div className="mb-1 text-xs text-(--system-gray-5)">
          Items marked with * are required
        </div>
        <ToggleButtonGroup
          isRawView={inputMode === 'raw'}
          onEnableRaw={() => setInputMode('raw')}
          onDisableRaw={() => setInputMode('pretty')}
        />
      </div>

      <div
        style={{
          height: 'calc(100vh - 295px)',
          width: '100%',
          border: formState.errors.inputs
            ? '1px solid var(--accent-red)'
            : 'none',
          borderRadius: formState.errors.inputs ? '4px' : '0',
        }}
        className="w-full"
      >
        {inputMode === 'pretty' ? (
          <>
            <Badge color="yellow">JSON forms are in beta</Badge>
            <SchemaForm
              formData={formDataValues}
              jsonSchema={jsonInputs}
              onChange={onRjsfChange}
              setState={({ errors, touchedId: id }) => {
                setRjsfFormState((prev) => ({
                  ...prev,
                  ...(id && {
                    touchedFields: { ...prev.touchedFields, [id]: true },
                  }),
                  ...(errors && { errors }),
                }))
              }}
              state={rjsfFormState}
              extraFormContext={datetimeFieldConfig}
            />
          </>
        ) : (
          <MonacoEditor
            beforeMount={registerFlyteLightTheme}
            height="100%"
            width="100%"
            defaultLanguage="json"
            onChange={(value) => {
              try {
                if (!value) return
                const parsedData = JSON.parse(value)
                setValue('formData', parsedData)
                clearErrors('inputs')
              } catch (e) {
                const message = e instanceof Error ? `: ${e.message}` : ''
                setError('inputs', {
                  message: `Invalid JSON${message}`,
                })
              }
            }}
            value={stringify(formDataValues, null, 2)}
            theme={FLYTE_LIGHT_THEME}
            options={{
              minimap: { enabled: false },
              readOnly: false,
              scrollBeyondLastLine: false,
              wordWrap: 'on',
              renderLineHighlight: 'none',
              overviewRulerLanes: 0,
              hideCursorInOverviewRuler: true,
              scrollbar: { vertical: 'auto', horizontal: 'auto' },
            }}
          />
        )}
      </div>
    </div>
  )
}
