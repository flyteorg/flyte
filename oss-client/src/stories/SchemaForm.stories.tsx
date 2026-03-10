import SchemaForm, { RjsfSchemaFormState } from '@/components/SchemaForm'
import { JSONSchema7, JSONSchema7Type, JSONSchema7TypeName } from 'json-schema'
import { FormState } from '@rjsf/core'
import { RJSFSchema, StrictRJSFSchema } from '@rjsf/utils'
import type { Meta, StoryObj } from '@storybook/nextjs'
import dayjs from 'dayjs'
import { useCallback, useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

type SchemaFormErrorWrapperType = {
  schema?: RJSFSchema
  formData?: FormState['formData']
}
const SchemaFormErrorWrapper: React.FC<SchemaFormErrorWrapperType> = ({
  schema,
  formData = {},
}) => {
  const methods = useForm<SchemaFormErrorWrapperType>({
    defaultValues: { schema, formData },
  })
  const [state, setState] = useState<RjsfSchemaFormState>({
    errors: [],
    touchedFields: {},
  })

  const onChange = useCallback(
    (data: FormState['formData']) => {
      methods.setValue('formData', data)
      methods.trigger('formData')
    },
    [methods],
  )

  return (
    <FormProvider {...methods}>
      <SchemaForm
        jsonSchema={methods.getValues('schema')}
        formData={methods.getValues('formData')}
        onChange={onChange}
        state={state}
        setState={({ errors, touchedId: id }) => {
          setState((prev) => ({
            ...prev,
            ...(id && {
              touchedFields: { ...prev.touchedFields, [id]: true },
            }),
            ...(errors && { errors }),
          }))
        }}
      />
    </FormProvider>
  )
}

type SampleValues = Partial<Record<JSONSchema7TypeName, JSONSchema7Type>>
type SchemaProps = StrictRJSFSchema['properties']

const sampleValues: SampleValues = {
  string: 'string',
  number: 1234,
  boolean: true,
  null: null,
}

function mkProps(): SchemaProps {
  return Object.entries(sampleValues).reduce(
    (props, [type, defVal]) => {
      return {
        ...props,
        [`sample_${type}`]: {
          title: `Sample ${type}`,
          type: type as JSONSchema7TypeName,
          default: defVal,
        },
      }
    },
    {} as StrictRJSFSchema['properties'],
  )
}

function mkSchema(
  allRequired?: boolean,
  nested?: JSONSchema7 & { key: string },
): StrictRJSFSchema {
  const properties = mkProps()

  return {
    type: 'object',
    properties: nested
      ? {
          ...properties,
          [nested.key]: nested,
        }
      : properties,
    required: allRequired ? Object.keys(properties ?? {}) : undefined,
  }
}

const meta: Meta<typeof SchemaForm> = {
  component: SchemaForm,
  tags: ['autodocs'],
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/design/8srmhAcZjvRfCgSGDvwbfZ/Union-V2?t=6ilJSZ3xK85DoHOO-0',
    },
  },
}

export default meta

type Story = StoryObj<typeof SchemaForm>

export const BasicReadonly: Story = {
  name: 'Basic (readonly)',
  args: {
    jsonSchema: mkSchema(),
    readonly: true,
  },
}

export const BasicReadonlyAllRequired: Story = {
  name: 'Basic (readonly) - all required',
  args: {
    jsonSchema: mkSchema(true),
    readonly: true,
  },
}

export const TextDifferentLength: Story = {
  name: 'Text different lengths',
  args: {
    jsonSchema: {
      readonly: false,
      type: 'object',
      properties: [1, 5, 10, 20, 30].reduce<SchemaProps>(
        (props, reps) => ({
          ...props,
          [`text_${reps}`]: {
            type: 'string',
            default: 'Lorem ipsum dolor sit amet'.repeat(reps),
          },
          [`no_spaces_${reps}`]: {
            type: 'string',
            default: 'CamelCaseTextThatWillRepeatMultipleTimes'.repeat(reps),
          },
        }),
        {},
      ),
    },
  },
}

export const DateAndTimeFields: Story = {
  name: 'Date and Time Fields',
  args: {
    readonly: false,
    jsonSchema: {
      type: 'object',
      title: 'Title',
      properties: {
        justDate: {
          type: 'string',
          format: 'date',
          title: 'Date',
          default: dayjs().format('YYYY-MM-DD'),
        },
        justTime: {
          type: 'string',
          format: 'time',
          title: 'Time',
          default: dayjs().format('HH:mm:ss'),
        },
        dateAndTime: {
          type: 'string',
          format: 'datetime',
          title: 'Datetime',
          default: dayjs().format('YYYY-MM-DD HH:mm:ss'),
        },
        rjsfWeirdness: {
          type: 'object',
          title: 'RJSF Weirdness: formats found in RJSF internals',
          properties: {
            altDate: {
              type: 'string',
              format: 'alt-date',
              title: '(Alt) date',
              default: dayjs().format('YYYY-MM-DD'),
            },
            altDatetime: {
              type: 'string',
              format: 'alt-datetime',
              title: '(Alt) datetime',
              default: dayjs().format('YYYY-MM-DD HH:mm:ss'),
            },
          },
        },
      },
    },
  },
}

export const NestedObject: Story = {
  name: 'Nested object',
  args: {
    readonly: true,
    jsonSchema: mkSchema(false, {
      ...mkSchema(false),
      key: 'nested',
      type: 'object',
    }),
  },
}
export const NestedObjectNestedRequired: Story = {
  name: 'Nested object with all required props',
  args: {
    readonly: true,
    jsonSchema: mkSchema(false, {
      ...mkSchema(true),
      key: 'nested',
      type: 'object',
    }),
  },
}

export const NestedArray: Story = {
  name: 'Nested array',
  args: {
    readonly: true,
    jsonSchema: {
      ...mkSchema(false, {
        type: 'array',
        key: 'nested_array',
        title: 'Nested array',
        items: { type: 'string', title: 'Nested val' },
        default: ['lorem ipsum', 'foo bar'],
      }),
    },
  },
}

export const NestedArrayOfObjects: Story = {
  name: 'Nested array of objects',
  args: {
    readonly: true,
    jsonSchema: {
      ...mkSchema(false, {
        type: 'array',
        key: 'nested_array',
        title: 'Nested array of objects',
        items: { ...mkSchema(), title: 'nested object' },
        default: [
          {
            sample_string: 'nested str',
            sample_number: 2137,
            sample_boolean: false,
            sample_null: null,
          },
          {
            sample_string: 'nested str 2',
            sample_number: 2138,
            sample_boolean: true,
            sample_null: null,
          },
        ],
      }),
    },
  },
}

export const NestedObjectWithArrayPropertyWithErrors: Story = {
  name: 'Nested object with array property with shown errors',
  args: {
    jsonSchema: {
      ...mkSchema(true, {
        properties: {
          regularProp: {
            type: 'string',
            default: 'string property',
          },
          arrayProp: {
            type: 'array',
            title: 'Array property of an object',
            items: { type: 'string' },
            default: ['string inside array inside object', 'and another one'],
          },
        },
        key: 'nested_object_with_array_property',
      }),
    },
  },
  render: (args) => <SchemaFormErrorWrapper schema={args.jsonSchema} />,
}

export const NestedObjectWithArrayOfObjectsProperty: Story = {
  name: 'Nested object with array of objects as a property',
  args: {
    readonly: true,
    jsonSchema: {
      ...mkSchema(false, {
        properties: {
          regularProp: {
            type: 'string',
            default: 'string property',
          },
          arrayProp: {
            type: 'array',
            title: 'Array property of an object',
            items: mkSchema(),
            default: [
              {
                sample_string: 'nested str',
                sample_number: 2137,
                sample_boolean: false,
                sample_null: null,
              },
              {
                sample_string: 'nested str 2',
                sample_number: 2138,
                sample_boolean: true,
                sample_null: null,
              },
            ],
          },
        },
        key: 'nested_object_with_array_property',
      }),
    },
  },
}
