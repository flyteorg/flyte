import { Button } from '@/components/Button'
import { INDENT_PER_LEVEL, MAX_INDENT } from '@/components/SchemaForm/constants'
import { useObjectPropertyContext } from '@/components/SchemaForm/CustomFields/Objects/ObjectPropertyContext'
import KeyNameInput from '@/components/SchemaForm/KeyNameInput'
import { MinusCircleIcon } from '@heroicons/react/24/outline'
import { FieldTemplateProps, RJSFValidationError } from '@rjsf/utils'
import React, { ReactElement, ReactNode, useMemo } from 'react'

const ThemedField: React.FC<{
  id?: string
  label: string
  type: string | string[] | undefined
  children: ReactNode
  required?: boolean
  errors?: RJSFValidationError[]
  hideErrors?: boolean
  schema?: FieldTemplateProps['schema']
  onKeyChange?: (key: string) => void
  onRemove?: (key: string) => () => void
}> = ({
  id,
  label,
  type,
  children,
  required,
  errors = [],
  hideErrors,
  schema,
  onKeyChange,
  onRemove,
}) => {
  const { parentName, index, depth: contextDepth } = useObjectPropertyContext()

  // Use context depth if available, otherwise assume first-level field (depth 1)
  // We can't reliably parse depth from id because field names may contain underscores
  // (e.g., "start_time" would be incorrectly parsed as two levels)
  const depth = contextDepth > 0 ? contextDepth : 1

  // Generate description ID for accessibility
  const descriptionId = id ? `${id}__description` : undefined

  // Add aria-describedby to children if description exists
  const childrenWithAria = useMemo(() => {
    if (!descriptionId || !schema?.description) {
      return children
    }

    // If children is a single React element, clone it and add aria-describedby
    if (React.isValidElement(children)) {
      const childElement = children as ReactElement<{
        'aria-describedby'?: string
        [key: string]: unknown
      }>
      const existingAriaDescribedBy = childElement.props?.['aria-describedby']
      const ariaDescribedBy = [existingAriaDescribedBy, descriptionId]
        .filter(Boolean)
        .join(' ')

      return React.cloneElement(childElement, {
        ...childElement.props,
        'aria-describedby': ariaDescribedBy || undefined,
      })
    }

    // If children is an array, try to clone the first element
    if (Array.isArray(children) && children.length > 0) {
      const firstChild = children[0]
      if (React.isValidElement(firstChild)) {
        const firstChildElement = firstChild as ReactElement<{
          'aria-describedby'?: string
          [key: string]: unknown
        }>
        const existingAriaDescribedBy =
          firstChildElement.props?.['aria-describedby']
        const ariaDescribedBy = [existingAriaDescribedBy, descriptionId]
          .filter(Boolean)
          .join(' ')

        return [
          React.cloneElement(firstChildElement, {
            ...firstChildElement.props,
            'aria-describedby': ariaDescribedBy || undefined,
          }),
          ...children.slice(1),
        ]
      }
    }

    return children
  }, [children, descriptionId, schema?.description])

  const allowChangeKey = !!schema?.__additional_property

  if (allowChangeKey) {
    return (
      <div className="flex w-full flex-col py-2" data-id={id}>
        <div className="my-2.5 flex flex-row items-center text-[14px] font-normal text-zinc-900 dark:text-white">
          <Button plain onClick={onRemove?.(label)} size="xxs" color="white">
            <MinusCircleIcon />
          </Button>
          {parentName}[{index}]
        </div>
        <div className="my-2 ml-12 flex flex-col">
          <div className="mb-2 flex flex-row items-center gap-x-1">
            <span className="text-[14px] font-semibold text-zinc-900 dark:text-white">
              key
            </span>
            <span className="flex items-center pl-0.5 text-[14px] tracking-[.25px] text-zinc-600 dark:text-(--system-gray-5)">
              (string)*
            </span>
          </div>
          <KeyNameInput label={label} type={type} onKeyChange={onKeyChange} />
        </div>
        <div className="ml-12 flex flex-col">
          <div className="mt-2 mb-2 flex flex-row gap-y-4">
            <span className="mr-1 text-[14px] font-semibold text-zinc-900 dark:text-white">
              value
            </span>
            <span className="flex items-center pl-0.5 text-[14px] tracking-[.25px] text-(--system-gray-5)">
              {type && `(${type})`}
              {required ? '*' : ''}
            </span>
          </div>
          {children}
        </div>
      </div>
    )
  }

  const indentPx = Math.min(depth * INDENT_PER_LEVEL, MAX_INDENT)
  // Only show border for depth > 1 (nested beyond first level)
  // First level fields (direct children of root properties) should not have borders
  const shouldShowBorder = depth > 1

  return (
    <div
      className={`grid w-full grid-cols-[1fr_2fr] border-b border-b-zinc-300 py-2 last:border-b-0 dark:border-b-zinc-800 ${
        shouldShowBorder
          ? 'border-l border-l-zinc-300/50 dark:border-l-zinc-700/50'
          : ''
      }`}
      style={{
        paddingLeft: indentPx > 0 ? `${indentPx}px` : undefined,
        paddingRight: 0,
        marginRight: 0,
      }}
    >
      <div className="flex flex-col items-start text-[14px] font-medium">
        <div className="flex items-center">
          {label}
          <span className="flex items-center pl-0.5 text-[14px] tracking-[.25px] text-zinc-600 dark:text-(--system-gray-5)">
            {type && `(${type})`}
            {required ? '*' : ''}
          </span>
        </div>
        {schema?.description && (
          <div
            id={descriptionId}
            className="mt-0.5 text-[13px] text-zinc-400 italic dark:text-zinc-500"
          >
            {schema.description}
          </div>
        )}
      </div>
      <div className="min-w-0">
        <div className="rjsf-form-field flex w-full min-w-0 text-sm/5 text-zinc-500">
          {childrenWithAria}
        </div>
        {errors.length > 0 && !hideErrors && (
          <div className="pt-1 pl-0.5 text-sm/5 text-red-400">
            <ul id={`${id}__error`}>
              {errors.map((e) => (
                <li key={e.name}>{e.message}</li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  )
}

export default ThemedField
