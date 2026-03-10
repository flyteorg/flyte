'use client'

import { Input, InputGroup } from '@/components/Input'
import { MagnifyingGlassIcon } from '@heroicons/react/16/solid'
import clsx from 'clsx'
import { Fragment, useMemo, useState } from 'react'
import { Controller, FieldValues, Path, useFormContext } from 'react-hook-form'
import { Field, Label } from './Fieldset'
import { ChevronDownIcon } from './icons/ChevronDownIcon'
import { Popover } from './Popovers'

export const SELECT_ALL_OPTIONS_VALUE = 'all-options-value'
export type SelectOption = {
  value: string
  label: string
  description?: string
}

export type SelectFieldProps<T extends FieldValues> = {
  name: Path<T>
  labelText: string
  options: SelectOption[]
  allOptionsLabel?: string
  disabled?: boolean
}

export function SelectField<T extends FieldValues>({
  name,
  labelText,
  options,
  allOptionsLabel,
  disabled = false,
}: SelectFieldProps<T>) {
  const { control } = useFormContext<T>()
  const [searchText, setSearchText] = useState('')
  const [isOpen, setIsOpen] = useState(false)

  const allOptions = useMemo(() => {
    return allOptionsLabel
      ? [
          { label: allOptionsLabel, value: SELECT_ALL_OPTIONS_VALUE },
          ...options,
        ]
      : options
  }, [allOptionsLabel, options])

  const filteredOptions = searchText
    ? allOptions.filter((option) =>
        option.label.toLowerCase().includes(searchText.toLowerCase()),
      )
    : allOptions

  return (
    <Field>
      <Label htmlFor={name}>{labelText}</Label>
      <Controller
        name={name}
        control={control}
        render={({ field }) => {
          return (
            <Popover
              placement="bottom-end"
              open={disabled ? false : isOpen}
              onOpenChange={(open) => !disabled && setIsOpen(open)}
              content={
                <div className="mt-1 w-full rounded bg-(--system-black) p-2 shadow sm:w-112">
                  <InputGroup className="mb-1 rounded-lg border-none sm:*:data-[slot=icon]:top-[9px] dark:bg-(--system-gray-3)">
                    <MagnifyingGlassIcon data-slot="icon" />
                    <Input
                      type="search"
                      hideBorder={true}
                      value={searchText}
                      onChange={(e) => setSearchText(e.target.value)}
                      placeholder="Search"
                      role="searchbox"
                    />
                  </InputGroup>

                  <ul
                    role="listbox"
                    aria-label={name}
                    className="max-h-60 w-full overflow-auto"
                  >
                    {filteredOptions.length === 0 ? (
                      <li
                        role="option"
                        aria-selected={false}
                        aria-disabled="true"
                        className="px-3 py-2 text-center text-sm text-gray-500"
                      >
                        No results found
                      </li>
                    ) : (
                      filteredOptions.map((option) => {
                        const isSelected = field.value === option.value
                        return (
                          <Fragment key={option.value}>
                            <li
                              role="option"
                              aria-selected={isSelected}
                              className={clsx(
                                'cursor-pointer rounded-lg px-1.5 py-1 text-sm hover:bg-(--system-gray-3) dark:text-(--system-gray-6) hover:dark:text-(--system-white)',
                                isSelected &&
                                  'bg-(--system-gray-2) text-(--system-white)',
                              )}
                              onClick={() => {
                                field.onChange(option.value)
                                setIsOpen(false)
                                setSearchText('')
                              }}
                            >
                              <div className="flex flex-col">
                                <span>{option.label}</span>
                                {option.description && (
                                  <span className="text-xs text-(--system-gray-5)">
                                    {option.description}
                                  </span>
                                )}
                              </div>
                            </li>
                            {option.value === SELECT_ALL_OPTIONS_VALUE && (
                              <hr className="my-1 w-full dark:text-white/15" />
                            )}
                          </Fragment>
                        )
                      })
                    )}
                  </ul>
                </div>
              }
            >
              <button
                type="button"
                disabled={disabled}
                className={clsx(
                  isOpen && 'outline-2 outline-offset-2 outline-[#6366F1]',
                  'mt-2 flex h-9.5 w-full items-center justify-between rounded-lg border border-zinc-950/10 px-3 py-1.5 text-sm sm:w-112 dark:border-zinc-700',
                  disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer',
                )}
                onClick={() => !disabled && setIsOpen((prev) => !prev)}
              >
                {allOptions.find((o) => o.value === field.value)?.label ?? (
                  <div />
                )}
                <ChevronDownIcon className="text-zinc-400" />
              </button>
            </Popover>
          )
        }}
      />
    </Field>
  )
}
