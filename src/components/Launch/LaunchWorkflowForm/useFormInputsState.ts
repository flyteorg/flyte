import { useDebouncedValue } from 'components/hooks/useDebouncedValue';
import { Core } from 'flyteidl';
import { useEffect, useState } from 'react';
import { validateInput } from './inputHelpers/inputHelpers';
import { InputProps, InputValue, ParsedInput } from './types';
import { convertFormInputsToLiterals } from './utils';

const debounceDelay = 500;

interface FormInputState extends InputProps {
    validate(): boolean;
}

interface FormInputsState {
    inputs: InputProps[];
    getValues(): Record<string, Core.ILiteral>;
    validate(): boolean;
}

function useFormInputState(parsedInput: ParsedInput): FormInputState {
    const [value, setValue] = useState<InputValue | undefined>(
        parsedInput.defaultValue
    );
    const [error, setError] = useState<string>();

    const validationValue = useDebouncedValue(value, debounceDelay);

    const validate = () => {
        const { name, required, typeDefinition } = parsedInput;
        try {
            validateInput({ name, required, typeDefinition, value });
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
        setValue(value);
    };

    return {
        ...parsedInput,
        error,
        onChange,
        validate,
        value,
        helperText: parsedInput.description
    };
}

/** Manages the state (value, error, validation) for a list of `ParsedInput` values.
 * NOTE: The input value for this hook is used to generate sub-hooks.
 * If the input value will change, the component using this hook should
 * be remounted (such as with the `key` prop) each time the value changes.
 * Otherwise we will end up calling the hooks for a component in a different order.
 * See https://reactjs.org/docs/hooks-rules.html#explanation
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
