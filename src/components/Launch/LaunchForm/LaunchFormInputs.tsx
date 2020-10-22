import * as React from 'react';
import { BlobInput } from './BlobInput';
import { CollectionInput } from './CollectionInput';
import { formStrings } from './constants';
import { LaunchState } from './launchMachine';
import { SimpleInput } from './SimpleInput';
import { useStyles } from './styles';
import {
    BaseInterpretedLaunchState,
    InputProps,
    InputType,
    LaunchFormInputsRef
} from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { UnsupportedRequiredInputsError } from './UnsupportedRequiredInputsError';
import { useFormInputsState } from './useFormInputsState';

function getComponentForInput(input: InputProps, showErrors: boolean) {
    const props = { ...input, error: showErrors ? input.error : undefined };
    switch (input.typeDefinition.type) {
        case InputType.Blob:
            return <BlobInput {...props} />;
        case InputType.Collection:
            return <CollectionInput {...props} />;
        case InputType.Map:
        case InputType.Unknown:
        case InputType.None:
            return <UnsupportedInput {...props} />;
        default:
            return <SimpleInput {...props} />;
    }
}

export interface LaunchFormInputsProps {
    state: BaseInterpretedLaunchState;
    variant: 'workflow' | 'task';
}

export const LaunchFormInputsImpl: React.RefForwardingComponent<
    LaunchFormInputsRef,
    LaunchFormInputsProps
> = ({ state, variant }, ref) => {
    const {
        parsedInputs,
        unsupportedRequiredInputs,
        showErrors
    } = state.context;
    const { getValues, inputs, validate } = useFormInputsState(parsedInputs);
    const styles = useStyles();
    React.useImperativeHandle(ref, () => ({
        getValues,
        validate
    }));

    const showInputs = [
        LaunchState.UNSUPPORTED_INPUTS,
        LaunchState.ENTER_INPUTS,
        LaunchState.VALIDATING_INPUTS,
        LaunchState.INVALID_INPUTS,
        LaunchState.SUBMIT_VALIDATING,
        LaunchState.SUBMITTING,
        LaunchState.SUBMIT_FAILED,
        LaunchState.SUBMIT_SUCCEEDED
    ].some(state.matches);

    return showInputs ? (
        <section title={formStrings.inputs}>
            {state.matches(LaunchState.UNSUPPORTED_INPUTS) ? (
                <UnsupportedRequiredInputsError
                    inputs={unsupportedRequiredInputs}
                    variant={variant}
                />
            ) : (
                <>
                    {inputs.map(input => (
                        <div key={input.label} className={styles.formControl}>
                            {getComponentForInput(input, showErrors)}
                        </div>
                    ))}
                </>
            )}
        </section>
    ) : null;
};

/** Renders an array of `ParsedInput` values using the appropriate
 * components. A `ref` to this component is used to access the current
 * form values and trigger manual validation if needed.
 */
export const LaunchFormInputs = React.forwardRef(LaunchFormInputsImpl);
