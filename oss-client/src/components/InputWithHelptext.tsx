import { Field, Label } from './Fieldset'
import { Input, InputProps } from './Input'

type InputWithHelpTextProps = InputProps & {
  helpText?: string
  id: string
  input?: React.ReactNode // optional for easier ref handling
  labelText: string
  placeholder?: string
}

export function InputWithHelpText({
  helpText,
  id,
  input,
  labelText,
  placeholder,
  ...props
}: InputWithHelpTextProps) {
  return (
    <Field>
      <Label htmlFor={id}>{labelText}</Label>
      {input || (
        <Input
          {...props}
          id={id}
          name={id}
          placeholder={placeholder}
          aria-describedby={`${id}-description`}
        />
      )}
      {helpText ? (
        <p
          data-slot="description"
          id={`${id}-description`}
          className="text-sm text-(--system-gray-5)"
        >
          {helpText}
        </p>
      ) : null}
    </Field>
  )
}
