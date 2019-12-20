import * as React from 'react';
import { CollectionInput } from './CollectionInput';
import { SimpleInput } from './SimpleInput';
import { useStyles } from './styles';
import {
    InputProps,
    InputType,
    LaunchWorkflowFormInputsRef,
    ParsedInput
} from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { useFormInputsState } from './useFormInputsState';

function getComponentForInput(input: InputProps) {
    switch (input.typeDefinition.type) {
        case InputType.Collection:
            return <CollectionInput {...input} />;
        case InputType.Map:
        case InputType.Schema:
        case InputType.Unknown:
        case InputType.None:
            return <UnsupportedInput {...input} />;
        default:
            return <SimpleInput {...input} />;
    }
}

export interface LaunchWorkflowFormInputsProps {
    inputs: ParsedInput[];
}

export const LaunchWorkflowFormInputsImpl: React.RefForwardingComponent<
    LaunchWorkflowFormInputsRef,
    LaunchWorkflowFormInputsProps
> = ({ inputs: parsedInputs }, ref) => {
    const { getValues, inputs, validate } = useFormInputsState(parsedInputs);
    const styles = useStyles();
    React.useImperativeHandle(ref, () => ({
        getValues,
        validate
    }));

    return (
        <>
            {inputs.map(input => (
                <div key={input.label} className={styles.formControl}>
                    {getComponentForInput(input)}
                </div>
            ))}
        </>
    );
};

/** Renders an array of `ParsedInput` values using the appropriate
 * components. A `ref` to this component is used to access the current
 * form values and trigger manual validation if needed.
 */
export const LaunchWorkflowFormInputs = React.forwardRef(
    LaunchWorkflowFormInputsImpl
);
