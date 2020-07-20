import * as React from 'react';
import { BlobInput } from './BlobInput';
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

function getComponentForInput(input: InputProps, showErrors: boolean) {
    const props = { ...input, error: showErrors ? input.error : undefined };
    switch (input.typeDefinition.type) {
        case InputType.Blob:
            return <BlobInput {...props} />;
        case InputType.Collection:
            return <CollectionInput {...props} />;
        case InputType.Map:
        case InputType.Schema:
        case InputType.Unknown:
        case InputType.None:
            return <UnsupportedInput {...props} />;
        default:
            return <SimpleInput {...props} />;
    }
}

export interface LaunchWorkflowFormInputsProps {
    inputs: ParsedInput[];
    showErrors?: boolean;
}

export const LaunchWorkflowFormInputsImpl: React.RefForwardingComponent<
    LaunchWorkflowFormInputsRef,
    LaunchWorkflowFormInputsProps
> = ({ inputs: parsedInputs, showErrors = true }, ref) => {
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
                    {getComponentForInput(input, showErrors)}
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
