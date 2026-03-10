import type { RJSFSchema } from '@rjsf/utils'
import { RegistryFieldsType } from '@rjsf/utils'
import DefaultObjectField from '@rjsf/core/lib/components/fields/ObjectField'
import React from 'react'
import CustomStringField from './CustomStringField'
import CustomNumberField from './CustomNumberField'
import CustomBlobField, { type BlobObjectFormData } from './CustomBlobField'

const BLOB_KEYS = ['dimensionality', 'format', 'uri']

function isBlobObjectSchema(schema: RJSFSchema): boolean {
  if (schema.type !== 'object' || schema.format !== 'blob') return false
  const props = schema.properties as Record<string, unknown> | undefined
  return (
    !!props && typeof props === 'object' && BLOB_KEYS.every((k) => k in props)
  )
}

const ObjectField: RegistryFieldsType['ObjectField'] = (props) => {
  const { schema, registry, formData } = props
  const resolved = (registry.schemaUtils.retrieveSchema(schema, formData) ??
    schema) as RJSFSchema
  if (isBlobObjectSchema(resolved)) {
    return (
      <CustomBlobField
        {...(props as Parameters<typeof CustomBlobField>[0])}
        schema={resolved}
      />
    )
  }
  return <DefaultObjectField {...props} />
}

const customFields: RegistryFieldsType = {
  NumberField: CustomNumberField,
  StringField: CustomStringField,
  ObjectField,
}

export { CustomStringField, CustomNumberField, CustomBlobField, customFields }
export type { BlobObjectFormData }
