import Form, { FormState, IChangeEvent } from '@rjsf/core'
import { FormContextType, RJSFSchema, RJSFValidationError } from '@rjsf/utils'
import validator from '@rjsf/validator-ajv8'
import type { JSONSchema7 } from 'json-schema'
import React, { useCallback, useMemo, useRef } from 'react'
import { customFields } from './CustomFields'
import ThemedForm from './ThemedForm'
import { normalizeSchemaDefs } from './utils'

export type RjsfSchemaFormState = {
  errors: RJSFValidationError[]
  touchedFields: Record<string, boolean>
}

export type SchemaFormProps = {
  jsonSchema?: RJSFSchema
  formData?: FormState['formData']
  readonly?: boolean
  onChange?: (formData: FormState['formData']) => void
  state?: RjsfSchemaFormState
  setState?: ({
    touchedId,
    errors,
  }: {
    touchedId?: string
    errors?: RJSFValidationError[]
  }) => void
  /** Additional context to pass to form fields (merged with internal context) */
  extraFormContext?: Record<string, unknown>
}

const SchemaForm: React.FC<SchemaFormProps> = ({
  jsonSchema,
  readonly = false,
  onChange,
  formData,
  setState,
  state,
  extraFormContext,
}) => {
  const { touchedFields, errors } = state ?? {}
  const formRef = useRef<Form<
    Record<string, unknown>,
    RJSFSchema,
    FormContextType
  > | null>(null)

  // Normalize schema by moving nested $defs to root level
  const normalizedSchema = useMemo(() => {
    if (!jsonSchema) {
      return jsonSchema
    }
    return normalizeSchemaDefs(jsonSchema as JSONSchema7)
  }, [jsonSchema])

  const handleOnChange = useCallback(
    ({ formData }: IChangeEvent) => {
      onChange?.(formData)
    },
    [onChange],
  )

  const handleOnBlur = useCallback(
    async (id: string) => {
      const isValid = await formRef.current?.validateForm()
      if (isValid) {
        setState?.({ errors: [], touchedId: id })
      } else {
        setState?.({ touchedId: id })
      }
    },
    [setState],
  )

  return (
    <ThemedForm
      ref={formRef}
      showErrorList={false}
      className="rjsf-form"
      schema={{ ...normalizedSchema, title: '' }}
      validator={validator}
      readonly={readonly}
      fields={customFields}
      liveValidate={false}
      onChange={handleOnChange}
      onBlur={handleOnBlur}
      onError={(errors) => setState?.({ errors })}
      formData={formData}
      formContext={{
        errors,
        touched: touchedFields ?? {},
        ...extraFormContext,
      }}
    >
      <></>
    </ThemedForm>
  )
}

export default SchemaForm
