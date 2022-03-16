import { useDebouncedValue } from 'components/hooks/useDebouncedValue';
import { Core } from 'flyteidl';
import { useEffect, useState } from 'react';
import { launchInputDebouncDelay } from './constants';
import {
  defaultValueForInputType,
  literalToInputValue,
  validateInput,
} from './inputHelpers/inputHelpers';
import { useInputValueCacheContext } from './inputValueCache';
import { InputProps, InputValue, ParsedInput } from './types';
import { convertFormInputsToLiterals, createInputCacheKey } from './utils';

interface FormInputState extends InputProps {
  validate(): boolean;
}

interface FormInputsState {
  inputs: InputProps[];
  getValues(): Record<string, Core.ILiteral>;
  validate(): boolean;
}

function useFormInputState(parsedInput: ParsedInput): FormInputState {
  const inputValueCache = useInputValueCacheContext();

  const cacheKey = createInputCacheKey(parsedInput.name, parsedInput.typeDefinition);

  const { initialValue, name, required, typeDefinition } = parsedInput;

  const [value, setValue] = useState<InputValue | undefined>(() => {
    const parsedInitialValue = initialValue
      ? literalToInputValue(parsedInput.typeDefinition, initialValue)
      : defaultValueForInputType(typeDefinition);
    return inputValueCache.has(cacheKey) ? inputValueCache.get(cacheKey) : parsedInitialValue;
  });
  const [error, setError] = useState<string>();

  const validationValue = useDebouncedValue(value, launchInputDebouncDelay);

  const validate = () => {
    try {
      validateInput({
        name,
        required,
        typeDefinition,
        value,
        initialValue: parsedInput.initialValue,
      });
      setError(undefined);
      return true;
    } catch (e) {
      setError((e as Error).message);
      return false;
    }
  };

  useEffect(() => {
    validate();
  }, [validationValue]);

  const onChange = (value: InputValue) => {
    inputValueCache.set(cacheKey, value);
    setValue(value);
  };

  return {
    ...parsedInput,
    error,
    initialValue,
    onChange,
    validate,
    value,
    helperText: parsedInput.description,
  };
}

/** Manages the state (value, error, validation) for a list of `ParsedInput` values.
 * NOTE: The input value for this hook is used to generate sub-hooks.
 * If the input value will change, the component using this hook should
 * be remounted (such as with the `key` prop) each time the value changes.
 * Otherwise we will end up calling the hooks for a component in a different order.
 * See https://reactjs.org/docs/hooks-rules.html#explanation
 */
export function useFormInputsState(parsedInputs: ParsedInput[]): FormInputsState {
  const inputs = parsedInputs.map(useFormInputState);

  const validate = () => {
    const valid = inputs.reduce((out, input) => out && input.validate(), true);
    return valid;
  };

  const getValues = () => convertFormInputsToLiterals(inputs);

  return {
    inputs,
    getValues,
    validate,
  };
}
