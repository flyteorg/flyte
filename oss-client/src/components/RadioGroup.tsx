import clsx from 'clsx'

export function RadioGroup<T extends string>({
  disabled,
  id,
  helpText,
  labelText,
  radioOptions,
  selectedOptionId,
  setSelectedOptionId,
}: {
  disabled?: boolean
  id: string
  helpText?: string | React.ReactNode
  labelText: string
  radioOptions: { id: T; title: string }[]
  selectedOptionId: string
  setSelectedOptionId: (newId: T) => void
}) {
  return (
    <fieldset>
      {labelText && (
        <legend className="text-sm/6 font-semibold text-(--system-white)">
          {labelText}
        </legend>
      )}
      {helpText && <p className="text-sm text-(--system-gray-5)">{helpText}</p>}

      <div className="mt-2 flex w-fit flex-row gap-2 rounded-sm bg-(--system-black) p-1">
        {radioOptions.map((radioOption) => {
          const isSelected = radioOption.id === selectedOptionId
          const uniqueId = `${id}-${radioOption.id}`
          return (
            <div key={radioOption.id}>
              <input
                disabled={disabled}
                id={uniqueId}
                name={id}
                onClick={() => setSelectedOptionId(radioOption.id)}
                type="radio"
                defaultChecked={!disabled && isSelected}
                className="peer sr-only cursor-none"
              />

              <label
                htmlFor={uniqueId}
                className={clsx(
                  'flex cursor-pointer items-center rounded-sm px-4 text-sm/6 font-medium text-(--system-gray-5) hover:text-(--system-white)',
                  // checked
                  'peer-checked:bg-(--system-gray-4) peer-checked:text-(--system-white)',
                )}
              >
                {radioOption.title}
              </label>
            </div>
          )
        })}
      </div>
    </fieldset>
  )
}
