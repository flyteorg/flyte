import { ErrorText } from '@/components/ErrorText'
import { Input, InputGroup } from '@/components/Input'
import { InputWithHelpText } from '@/components/InputWithHelptext'
import { RadioGroup } from '@/components/RadioGroup'
import { RunService } from '@/gen/flyteidl2/workflow/run_service_pb'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'
import { useOrg } from '@/hooks/useOrg'
import { getRunIdentifier } from '@/hooks/useRunDetails'
import { is404Error } from '@/lib/errorUtils'
import { useParams } from 'next/navigation'
import { useEffect, useRef } from 'react'
import { useFormContext } from 'react-hook-form'
import { TabLayout } from './LaunchFormTabs'
import {
  LaunchFormState,
  RUN_NAME_MAX_LENGTH,
  RUN_NAME_MIN_LENGTH,
  RUN_NAME_PATTERN,
  RUN_NAME_PATTERN_DESCRIPTION,
} from './types'

type OverwriteCache = 'overwrite' | 'not-overwrite'

const overwriteCacheValues: Record<OverwriteCache, boolean | undefined> = {
  overwrite: true,
  'not-overwrite': false,
}

type ForceInterruptibleEnable = 'enabled' | 'disabled' | 'not set'

const forceInterruptibleValues: Record<
  ForceInterruptibleEnable,
  boolean | undefined
> = {
  enabled: true,
  disabled: false,
  'not set': undefined,
}

const RUN_NAME_DEBOUNCE_MS = 500

function runNameSyncError(value: string | undefined): string | true {
  const str = value ?? ''
  if (str.length === 0) return true
  if (str !== str.trim()) {
    return 'Run name cannot have leading or trailing spaces'
  }
  const trimmed = str.trim()
  if (trimmed.length < RUN_NAME_MIN_LENGTH) {
    return `Run name must be at least ${RUN_NAME_MIN_LENGTH} characters`
  }
  if (!RUN_NAME_PATTERN.test(trimmed)) {
    return RUN_NAME_PATTERN_DESCRIPTION
  }
  return true
}

export const LaunchFormSettings = () => {
  const {
    clearErrors,
    formState: { errors },
    register,
    setValue,
    watch,
    setError,
  } = useFormContext<LaunchFormState>()
  const params = useParams<{ domain?: string; project?: string }>()
  const org = useOrg()
  const { triggerName } = useLaunchFormState()
  const runClient = useConnectRpcClient(RunService)
  const runName = watch('runName')
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const runNameRef = useRef(runName)
  runNameRef.current = runName

  const project = triggerName ? triggerName.project : (params.project ?? '')
  const domain = triggerName ? triggerName.domain : (params.domain ?? '')

  // Sync from form so rehydration from previous run is reflected in the UI
  const overwriteCache = watch('overwriteCache')
  const interruptible = watch('interruptible')
  const overwriteCacheValue: OverwriteCache =
    overwriteCache === true ? 'overwrite' : 'not-overwrite'
  const forceInterruptibleValue: ForceInterruptibleEnable =
    interruptible === true
      ? 'enabled'
      : interruptible === false
        ? 'disabled'
        : 'not set'

  // Debounced async check: run name already exists
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
      debounceRef.current = null
    }
    const str = (runName ?? '').trim()
    if (str.length === 0) {
      clearErrors('runName')
      return
    }
    if (runNameSyncError(runName) !== true) return
    if (!project || !domain) return

    debounceRef.current = setTimeout(async () => {
      debounceRef.current = null
      try {
        const runId = getRunIdentifier({
          org,
          project,
          domain,
          name: str,
        })
        await runClient.getRunDetails({ runId })
        if ((runNameRef.current ?? '').trim() === str) {
          setError('runName', {
            type: 'validation',
            message: 'A run with this name already exists',
          })
        }
      } catch (error) {
        if ((runNameRef.current ?? '').trim() !== str) return
        if (is404Error(error)) {
          clearErrors('runName')
        } else {
          console.error('Run name availability check failed:', error)
          setError('runName', {
            type: 'validation',
            message: 'Could not verify run name; please try again',
          })
        }
      }
    }, RUN_NAME_DEBOUNCE_MS)

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [runName, project, domain, org, runClient, setError, clearErrors])

  return (
    <TabLayout>
      <div className="pt-3 pb-5 text-sm text-(--system-gray-5)">
        These settings will override the default run behavior
      </div>

      <InputWithHelpText
        id="run-name"
        input={
          <>
            <Input
              {...(errors.runName ? { 'data-invalid': true } : {})}
              aria-invalid={!!errors.runName}
              aria-describedby={
                errors.runName
                  ? 'run-name-description run-name-error'
                  : 'run-name-description'
              }
              {...register('runName', {
                required: false,
                maxLength: {
                  value: RUN_NAME_MAX_LENGTH,
                  message: `Run name must be at most ${RUN_NAME_MAX_LENGTH} characters`,
                },
                validate: (value) => runNameSyncError(value),
                onChange: (e) => {
                  if (errors.runName && e.target.value) clearErrors('runName')
                },
              })}
            />
            {errors.runName ? (
              <span id="run-name-error" role="alert">
                <ErrorText>{errors.runName.message}</ErrorText>
              </span>
            ) : null}
          </>
        }
        labelText="Run name"
        helpText={`String used to identify this run. If provided, ${RUN_NAME_MIN_LENGTH}–${RUN_NAME_MAX_LENGTH} characters; only letters, numbers, hyphens, and underscores (no spaces). If not specified, the name will be auto generated.`}
      />

      <InputGroup className="py-7">
        <RadioGroup<OverwriteCache>
          id="overwrite-cache"
          labelText="Overwrite cached outputs"
          helpText="If overwrite is selected, this run will overwrite previously-computed cached outputs."
          radioOptions={[
            { id: 'overwrite', title: 'Overwrite' },
            { id: 'not-overwrite', title: 'Do not overwrite' },
          ]}
          selectedOptionId={overwriteCacheValue}
          setSelectedOptionId={(newValue: OverwriteCache) => {
            const stateValue = overwriteCacheValues[newValue]
            setValue('overwriteCache', stateValue ?? false)
          }}
        />
      </InputGroup>

      <InputGroup>
        <RadioGroup<ForceInterruptibleEnable>
          id="enable-force-interruptible"
          labelText="Force interruptible"
          helpText={
            <>
              Overrides the interruptible flag of a task for this run, allowing
              it to be forced on or off. Select <strong>Use default</strong> to
              preserve the task’s default behavior.
            </>
          }
          radioOptions={[
            { id: 'enabled', title: 'Enabled' },
            { id: 'disabled', title: 'Disabled' },
            { id: 'not set', title: 'Use default' },
          ]}
          selectedOptionId={forceInterruptibleValue}
          setSelectedOptionId={(newValue: ForceInterruptibleEnable) => {
            const stateValue = forceInterruptibleValues[newValue]
            setValue('interruptible', stateValue)
          }}
        />
      </InputGroup>
      {/*
      <InputWithHelpText
        id="service-account"
        input={<Input {...register('serviceAccount')} />}
        labelText="Service account"
        helpText={`The service account to use for this run eg “default-service-account.”`}
      /> */}
    </TabLayout>
  )
}
