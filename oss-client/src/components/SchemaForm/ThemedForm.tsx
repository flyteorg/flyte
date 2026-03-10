import CustomObjectField from '@/components/SchemaForm/CustomFields/CustomObjectField'
import { withTheme } from '@rjsf/core'
import UnsupportedField from '@rjsf/core/lib/components/templates/UnsupportedField'
import { RJSFValidationError } from '@rjsf/utils'
import ThemedField from './ThemedField'
import {
  CollectionAdd,
  CollectionMoveDown,
  CollectionMoveUp,
  CollectionRemove,
} from './Widgets/Buttons'
import { FIELD_NAME_PREFIX } from './CustomFields/hooks/useHtmlElementProps'

const ThemedForm = withTheme({
  templates: {
    UnsupportedFieldTemplate: (props) => {
      if (props.schema.type) {
        // fallback
        return <UnsupportedField {...props} />
      }
    },
    ButtonTemplates: {
      AddButton: CollectionAdd,
      RemoveButton: CollectionRemove,
      MoveUpButton: CollectionMoveUp,
      MoveDownButton: CollectionMoveDown,
    },
    ObjectFieldTemplate: CustomObjectField,
    FieldTemplate: ({
      id,
      children,
      label,
      displayLabel,
      schema,
      required,
      formContext,
      onKeyChange,
      onDropPropertyClick,
    }) => {
      const fieldErrors = formContext?.errors?.filter(
        (e: RJSFValidationError) => {
          const propertyName = e.property?.startsWith('.')
            ? e.property.slice(1)
            : e.property
          return FIELD_NAME_PREFIX + propertyName?.replaceAll('.', '_') === id
        },
      )
      const displayType = schema.format === 'blob' ? 'file' : schema.type
      const showLabelAndLayout =
        displayLabel || schema.type === 'boolean' || schema.format === 'blob'
      return showLabelAndLayout ? (
        <ThemedField
          id={id}
          label={label}
          type={displayType}
          schema={schema}
          required={required}
          errors={fieldErrors}
          hideErrors={!formContext?.touched[id]}
          onKeyChange={onKeyChange}
          onRemove={onDropPropertyClick}
        >
          {children}
        </ThemedField>
      ) : (
        children
      )
    },
  },
})

export default ThemedForm
