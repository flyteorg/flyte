import clsx from 'clsx'
import { useFormContext, Controller } from 'react-hook-form'

import { Input } from '@/components/Input'
import { Field, Label } from '@/components/Fieldset'
import { RadioGroup } from '@/components/RadioGroup'
import { AutomationTypeFields, automationTypes } from './AutomationTypes'
import { CreateTriggerState } from './types'

export const TriggerDefinition = () => {
  const {
    control,
    register,
    watch,
    formState: { errors, touchedFields, isSubmitted },
  } = useFormContext<CreateTriggerState>()
  const automationType = watch('automationType')

  const showNameError = !!errors.name && (touchedFields.name || isSubmitted)

  return (
    <div className="flex flex-col gap-2">
      <Field>
        <Label className={clsx(showNameError && '!text-(--accent-red)')}>
          Name
        </Label>
        <Input
          id="triggerName"
          {...register('name', { required: 'Name is required' })}
          className={clsx(showNameError && '!border-(--accent-red)')}
        />
        {showNameError && (
          <div className="mt-1 text-sm text-(--accent-red)">
            {errors.name?.message?.toString() ?? 'Name is required'}
          </div>
        )}
      </Field>
      <Field>
        <Label>
          Description{' '}
          <span className="font-normal dark:text-(--system-gray-5)">
            (Optional)
          </span>
        </Label>
        <Input id="triggerDescription" {...register('description')} />
      </Field>
      <Controller
        name="automationType"
        control={control}
        defaultValue={
          automationTypes[0].id === 'cron-schedule'
            ? 'cron-schedule'
            : 'fixed-rate'
        }
        render={({ field }) => (
          <RadioGroup
            id="automation-type"
            labelText="Automation type"
            radioOptions={automationTypes}
            selectedOptionId={field.value}
            setSelectedOptionId={field.onChange}
          />
        )}
      />
      <AutomationTypeFields automationType={automationType} />
    </div>
  )
}
