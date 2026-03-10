import clsx from 'clsx'
import { Controller, useFormContext, useWatch } from 'react-hook-form'
import { Field, Label } from '@/components/Fieldset'
import { Checkbox } from '@/components/Checkbox'
import { Input } from '@/components/Input'
import { FixedRateTrigger } from './types'

export const FixedRateInput = () => {
  const {
    control,
    getValues,
    setValue,
    formState: { errors, touchedFields, isSubmitted },
  } = useFormContext<FixedRateTrigger>()

  const shouldStartImmediately =
    useWatch({ control, name: 'shouldStartImmediately' }) ?? false

  const handleCheckboxChange = (newValue: boolean) => {
    // Mutually exclusive with manual start time.
    // If the user chooses "Start immediately", we clear the manual value and disable the input.
    setValue('shouldStartImmediately', newValue, {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })

    if (newValue) {
      setValue('startTime', '', {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: true,
      })
    }
  }

  return (
    <div className="gap1 flex flex-col gap-1">
      <Field>
        <Label
          className={clsx(
            !!errors.interval &&
              (touchedFields.interval || isSubmitted) &&
              '!text-(--accent-red)',
          )}
        >
          Interval
        </Label>
        <Controller
          name="interval"
          rules={{
            required: 'Interval is required',
            min: { value: 1, message: 'Interval must be at least 1' },
            validate: (value) => {
              const n = typeof value === 'number' ? value : Number(value)
              return (
                (Number.isFinite(n) && n > 0) ||
                'Interval must be a positive number'
              )
            },
          }}
          render={({ field }) => (
            <Input
              type="number"
              autoComplete="off"
              className={clsx(
                !!errors.interval &&
                  (touchedFields.interval || isSubmitted) &&
                  '!border-(--accent-red)',
              )}
              value={field.value ?? ''}
              onChange={(e) => field.onChange(e.currentTarget.valueAsNumber)}
              onBlur={field.onBlur}
              name={field.name}
              ref={field.ref}
            />
          )}
        />
        <div className="flex justify-between text-sm text-zinc-500">
          {typeof errors.interval?.message === 'string' &&
          (touchedFields.interval || isSubmitted) ? (
            <div className="text-(--accent-red)">{errors.interval.message}</div>
          ) : (
            'Input an interval in minutes'
          )}
        </div>
      </Field>

      <Field>
        <Label
          className={clsx(
            !!errors.startTime &&
              (touchedFields.startTime || isSubmitted) &&
              '!text-(--accent-red)',
          )}
        >
          Start time
        </Label>
        <Controller
          name="startTime"
          rules={{
            validate: (value) => {
              const immediate = getValues('shouldStartImmediately')

              // Require at least one of: a manual start time OR "Start immediately".
              if (!value && !immediate) {
                return 'Select a start time'
              }

              // If no manual time, and immediate is checked, that's valid.
              if (!value) return true

              const selected = new Date(value).getTime()
              const now = Date.now()
              return selected >= now || 'Start time cannot be in the past'
            },
          }}
          render={({ field }) => (
            <Input
              type="datetime-local"
              {...field}
              disabled={shouldStartImmediately}
              onChange={(e) => {
                // If the user starts setting a manual time, turn off "Start immediately".
                if (shouldStartImmediately) {
                  setValue('shouldStartImmediately', false, {
                    shouldDirty: true,
                    shouldTouch: true,
                    shouldValidate: true,
                  })
                }
                field.onChange(e)
              }}
              className={clsx(
                !!errors.startTime &&
                  (touchedFields.startTime || isSubmitted) &&
                  '!border-(--accent-red)',
              )}
            />
          )}
        />
        <div className="flex justify-between text-sm text-zinc-500">
          {typeof errors.startTime?.message === 'string' &&
          (touchedFields.startTime || isSubmitted) ? (
            <div className="text-(--accent-red)">
              {errors.startTime.message}
            </div>
          ) : (
            'Select a start time (optional, must be in the future).'
          )}
        </div>
      </Field>

      <Field className="flex items-center gap-1">
        <Controller
          name="shouldStartImmediately"
          render={({ field }) => (
            <Checkbox
              {...field}
              name="shouldStartImmediately"
              checked={!!field.value}
              onChange={(next: boolean) => {
                field.onChange(next)
                handleCheckboxChange(next)
              }}
            />
          )}
        />

        <Label className="!text-zinc-500">Start immediately</Label>

        {typeof errors.shouldStartImmediately?.message === 'string' &&
        (touchedFields.shouldStartImmediately || isSubmitted) ? (
          <div className="text-sm text-(--accent-red)">
            {errors.shouldStartImmediately.message}
          </div>
        ) : null}
      </Field>
    </div>
  )
}
