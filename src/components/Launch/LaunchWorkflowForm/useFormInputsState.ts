import { useDebouncedValue } from 'components/hooks/useDebouncedValue';
import { Core } from 'flyteidl';
import { useEffect, useState } from 'react';
import { validateInput } from './inputHelpers/inputHelpers';
import { InputProps, InputValue, ParsedInput } from './types';
import { convertFormInputsToLiterals } from './utils';

// TODO: Maybe make this adjustable based on type
const debounceDelay = 500;

interface FormInputState extends InputProps {
    validate(): boolean;
}

interface FormInputsState {
    inputs: InputProps[];
    getValues(): Record<string, Core.ILiteral>;
    validate(): boolean;
}

// TODO: Mark values valid/invalid on initial setup so that manual
// validation isn't required later. Required values
// should default to being invalid.
function useFormInputState(parsedInput: ParsedInput): FormInputState {
    const [value, setValue] = useState<InputValue>();
    const [error, setError] = useState<string>();

    const validationValue = useDebouncedValue(value, debounceDelay);

    const validate = () => {
        const { typeDefinition } = parsedInput;
        try {
            validateInput({ name, typeDefinition, value });
            setError(undefined);
            return true;
        } catch (validationError) {
            setError(`${validationError}`);
            return false;
        }
    };

    useEffect(() => {
        validate();
    }, [validationValue]);

    return {
        ...parsedInput,
        error,
        validate,
        value,
        helperText: parsedInput.description,
        onChange: setValue
    };
}

/** Manages the state (value, error, validation) for a list of `ParsedInput` values.
 * NOTE: The input value for this hook is used to generate sub-hooks.
 * If the input value will change, the component using this hook should
 * be remounted (such as with the `key` prop) each time the value changes.
 */
export function useFormInputsState(
    parsedInputs: ParsedInput[]
): FormInputsState {
    const inputs = parsedInputs.map(useFormInputState);

    const validate = () => {
        const valid = inputs.reduce(
            (out, input) => out && input.validate(),
            true
        );
        return valid;
    };

    const getValues = () => convertFormInputsToLiterals(inputs);

    return {
        inputs,
        getValues,
        validate
    };
}
