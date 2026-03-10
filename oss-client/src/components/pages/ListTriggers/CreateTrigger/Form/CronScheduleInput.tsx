import { useFormContext, Controller } from 'react-hook-form'
import clsx from 'clsx'
import cronstrue from 'cronstrue'
import cronParser from 'cron-parser'
import { ArrowUpRightIcon } from '@heroicons/react/20/solid'
import { useMemo, useState } from 'react'
import { Field, Label } from '@/components/Fieldset'
import { Input } from '@/components/Input'
import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { getBrowserTimezone, getTimeZoneOptions } from './getTimezones'
import { SearchBar } from '@/components/SearchBar'
import { Link } from '@/components/Link'

const parseCron = (expr: string) => {
  let string = ''
  let isValid = true
  let error = ''
  try {
    cronParser.parse(expr)
    string = cronstrue.toString(expr)
  } catch (e) {
    isValid = false
    if (e instanceof Error) {
      error = e.message
    }
  }
  return { isValid, human: string, error: error }
}

export const CronScheduleInput = () => {
  const {
    control,
    watch,
    setValue,
    formState: { errors, touchedFields, isSubmitted },
  } = useFormContext()

  const cronExpression = watch('cronExpression')
  const [timezoneInput, setTimezoneInput] = useState('')

  const { isValid, human, error } = useMemo(() => {
    return cronExpression
      ? parseCron(cronExpression)
      : { isValid: false, human: '', error: '' }
  }, [cronExpression])

  const options: MenuItem[] = useMemo(() => {
    const search = {
      id: 'search',
      type: 'custom' as const,
      component: (
        <SearchBar
          value={timezoneInput}
          onChange={(e) => {
            setTimezoneInput(e.target.value)
          }}
        />
      ),
    }

    const timezones: MenuItem[] = getTimeZoneOptions()
      .map((option: string) => ({
        id: option,
        label: option,
        onClick: () => setValue('timezone', option),
      }))
      .filter((tz) =>
        tz.label
          .toLocaleLowerCase()
          .includes(timezoneInput.toLocaleLowerCase()),
      )
    return [search, ...timezones]
  }, [timezoneInput, setValue])

  const showError =
    !!errors.cronExpression && (touchedFields.cronExpression || isSubmitted)

  return (
    <div className="flex flex-col gap-2">
      <Field>
        <Label className={clsx(showError && '!text-(--accent-red)')}>
          Cron Expression
        </Label>
        <Controller
          name="cronExpression"
          control={control}
          defaultValue={'* * * * *'}
          rules={{
            required: 'Cron expression is required',
            validate: (value) => {
              const { isValid } = parseCron(value)
              return isValid || 'Invalid cron expression'
            },
          }}
          render={({ field }) => (
            <Input
              autoComplete="off"
              className={clsx(showError && '!border-(--accent-red)')}
              {...field}
              placeholder="* * * * *"
            />
          )}
        />
        <div className="flex justify-between text-sm text-zinc-500">
          {cronExpression ? (
            <div className={clsx(showError && 'text-(--accent-red)')}>
              {showError
                ? typeof errors.cronExpression?.message === 'string'
                  ? errors.cronExpression.message
                  : 'Invalid cron expression'
                : isValid
                  ? human
                  : error || 'Invalid cron expression'}
            </div>
          ) : (
            'Input a cron expression'
          )}
          <Link
            href="https://crontab.cronhub.io/"
            className="flex items-center gap-1 text-sm text-zinc-500"
            target="_blank"
          >
            Cron expression generator
            <ArrowUpRightIcon
              data-slot="icon"
              className="size-3 dark:text-zinc-500"
            />
          </Link>
        </div>
      </Field>
      <Field className="flex flex-col">
        <Label>Timezone</Label>
        <Controller
          name="timezone"
          control={control}
          defaultValue={getBrowserTimezone()}
          render={({ field }) => (
            <PopoverMenu
              items={options}
              label={field.value}
              menuClassName="overflow-y-auto max-h-60"
              triggerClassName="border-1 dark:border-(--system-gray-4) w-[fit-content]"
            />
          )}
        />
      </Field>
    </div>
  )
}
