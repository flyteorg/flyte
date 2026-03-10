import { Controller, useFormContext } from 'react-hook-form'
import { InputGroup } from '@/components/Input'
import { RadioGroup } from '@/components/RadioGroup'
import { CreateTriggerState } from './types'
import { InputWithHelpText } from '@/components/InputWithHelptext'

type ForceInterruptibleEnable = 'enabled' | 'disabled' | 'not set'

const forceInterruptibleValues: Record<
  ForceInterruptibleEnable,
  boolean | undefined
> = {
  enabled: true,
  disabled: false,
  'not set': undefined,
}

type OverwriteCache = 'overwrite' | 'not-overwrite'

const overwriteCacheValues: Record<OverwriteCache, boolean | undefined> = {
  overwrite: true,
  'not-overwrite': false,
}

export const TriggerAdvancedInputs = () => {
  const { control, register } = useFormContext<CreateTriggerState>()

  return (
    <div className="p-2">
      <InputGroup>
        <Controller
          name="interruptible"
          control={control}
          defaultValue={undefined}
          render={({ field }) => {
            const selectedOptionId: ForceInterruptibleEnable =
              field.value === true
                ? 'enabled'
                : field.value === false
                  ? 'disabled'
                  : 'not set'

            return (
              <RadioGroup<ForceInterruptibleEnable>
                id="enable-force-interruptible"
                labelText="Force interruptible"
                helpText={
                  <>
                    Overrides the interruptible flag of a task for this run,
                    allowing it to be forced on or off. Select{' '}
                    <strong>Use default</strong> to preserve the task’s default
                    behavior.
                  </>
                }
                radioOptions={[
                  { id: 'enabled', title: 'Enabled' },
                  { id: 'disabled', title: 'Disabled' },
                  { id: 'not set', title: 'Use default' },
                ]}
                selectedOptionId={selectedOptionId}
                setSelectedOptionId={(newValue: ForceInterruptibleEnable) => {
                  field.onChange(forceInterruptibleValues[newValue])
                }}
              />
            )
          }}
        />
      </InputGroup>

      <InputGroup className="py-7">
        <Controller
          name="overwriteCache"
          control={control}
          defaultValue={
            false as unknown as CreateTriggerState['overwriteCache']
          }
          render={({ field }) => {
            const selectedOptionId: OverwriteCache =
              field.value === false ? 'not-overwrite' : 'overwrite'

            return (
              <RadioGroup<OverwriteCache>
                id="overwrite-cache"
                labelText="Overwrite cached outputs"
                helpText="If overwrite is selected, this run will overwrite previously-computed cached outputs."
                radioOptions={[
                  { id: 'overwrite', title: 'Overwrite' },
                  { id: 'not-overwrite', title: 'Do not overwrite' },
                ]}
                selectedOptionId={selectedOptionId}
                setSelectedOptionId={(newValue: OverwriteCache) => {
                  const next = overwriteCacheValues[newValue]
                  field.onChange(Boolean(next))
                }}
              />
            )
          }}
        />
      </InputGroup>

      <InputWithHelpText
        id="service-account"
        labelText="Service account"
        {...register('serviceAccount')}
      />
      <InputWithHelpText
        id="dataOutputLocation"
        labelText="Data output location"
        {...register('dataOutputLocation')}
      />

      <InputGroup className="py-7">
        <Controller
          name="activeOnCreation"
          control={control}
          defaultValue={
            true as unknown as CreateTriggerState['activeOnCreation']
          }
          render={({ field }) => {
            const selectedOptionId: 'true' | 'false' =
              field.value === false ? 'false' : 'true'
            return (
              <RadioGroup<'true' | 'false'>
                id="activate"
                labelText="Activate on creation"
                helpText="Trigger will be active immediately after creation"
                radioOptions={[
                  { id: 'true', title: 'True' },
                  { id: 'false', title: 'False' },
                ]}
                selectedOptionId={selectedOptionId}
                setSelectedOptionId={(newValue) => {
                  field.onChange(newValue === 'true')
                }}
              />
            )
          }}
        />
      </InputGroup>
    </div>
  )
}
