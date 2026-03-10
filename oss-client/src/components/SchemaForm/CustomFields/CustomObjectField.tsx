import { ObjectFieldTemplateProps } from '@rjsf/utils'
import React, { useId, useMemo } from 'react'
import ObjectPropertyAdd from '@/components/SchemaForm/Widgets/Buttons/ObjectPropertyAdd'
import {
  ObjectPropertyContextProvider,
  useObjectPropertyContext,
} from '@/components/SchemaForm/CustomFields/Objects/ObjectPropertyContext'
import { INDENT_PER_LEVEL, MAX_INDENT } from '@/components/SchemaForm/constants'

const CustomObjectField: React.FC<ObjectFieldTemplateProps> = ({
  title,
  schema,
  properties,
  required,
  readonly,
  onAddClick,
  ...otherProps
}) => {
  const { depth: contextDepth } = useObjectPropertyContext()

  // For objects with a title, they represent a nested level, so increment depth
  // For root level objects (no title), use the context depth as is
  const currentDepth = title ? contextDepth + 1 : contextDepth
  const allowChangeKey = !!schema?.additionalProperties && !readonly
  const typeLabel = useMemo(
    () =>
      schema.type
        ? Array.isArray(schema.type)
          ? schema.type.join(', ')
          : schema.type
        : null,
    [schema.type],
  )

  // Generate a unique ID for the description element for accessibility
  const descriptionId = useId()

  // Add indentation for nested objects
  // Only indent if this is a nested object (has a title)
  const indentPx = title
    ? Math.min(currentDepth * INDENT_PER_LEVEL, MAX_INDENT)
    : 0
  const hasNestedContent = properties.length > 0
  const isNested = title && currentDepth > 0

  return (
    <div
      className={`flex w-full flex-col border-b border-b-zinc-300 py-2 dark:border-b-zinc-800 [&:last-child]:border-b-0 ${
        isNested
          ? 'border-l border-l-zinc-300/50 dark:border-l-zinc-700/50'
          : ''
      }`}
      style={{
        paddingLeft: isNested ? `${indentPx}px` : undefined,
        paddingRight: 0,
        marginRight: 0,
      }}
    >
      {title && (
        <div className="flex flex-col text-zinc-900 dark:text-white">
          <div className="flex flex-row">
            <span className="text-[14px] font-medium">{title}</span>
            <span className="flex items-center pl-0.5 text-[14px] tracking-[.25px] text-zinc-600 dark:text-(--system-gray-5)">
              {typeLabel && `(${typeLabel})`} {required ? '*' : ''}
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
      )}
      {hasNestedContent && (
        <div className="mt-2 flex flex-col">
          {properties.map(({ content }, index) => (
            <ObjectPropertyContextProvider
              key={index}
              parentName={title || ''}
              index={index}
              depth={currentDepth}
              objectDescriptionId={
                schema?.description ? descriptionId : undefined
              }
            >
              {content}
            </ObjectPropertyContextProvider>
          ))}
        </div>
      )}
      <div className="flex flex-row items-center justify-start py-4">
        {allowChangeKey && (
          <ObjectPropertyAdd
            {...otherProps}
            propertyName={title}
            propertyCount={properties.length}
            onClick={onAddClick(schema)}
          />
        )}
      </div>
    </div>
  )
}

export default CustomObjectField
