import { Typography } from '@material-ui/core';
import * as React from 'react';
import { BlobInput } from './BlobInput';
import { CollectionInput } from './CollectionInput';
import { formStrings, inputsDescription } from './constants';
import { LaunchState } from './launchMachine';
import { NoInputsNeeded } from './NoInputsNeeded';
import { SimpleInput } from './SimpleInput';
import { StructInput } from './StructInput';
import { useStyles } from './styles';
import { BaseInterpretedLaunchState, InputProps, InputType, LaunchFormInputsRef } from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { UnsupportedRequiredInputsError } from './UnsupportedRequiredInputsError';
import { useFormInputsState } from './useFormInputsState';
import { isEnterInputsState } from './utils';

function getComponentForInput(input: InputProps, showErrors: boolean) {
  const props = { ...input, error: showErrors ? input.error : undefined };
  switch (input.typeDefinition.type) {
    case InputType.Blob:
      return <BlobInput {...props} />;
    case InputType.Collection:
      return <CollectionInput {...props} />;
    case InputType.Struct:
      return <StructInput {...props} />;
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

const RenderFormInputs: React.FC<{
  inputs: InputProps[];
  showErrors: boolean;
  variant: LaunchFormInputsProps['variant'];
}> = ({ inputs, showErrors, variant }) => {
  const styles = useStyles();
  return inputs.length === 0 ? (
    <NoInputsNeeded variant={variant} />
  ) : (
    <>
      <header className={styles.sectionHeader}>
        <Typography variant="h6">{formStrings.inputs}</Typography>
        <Typography variant="body2">{inputsDescription}</Typography>
      </header>
      {inputs.map((input) => (
        <div key={input.label} className={styles.formControl}>
          {getComponentForInput(input, showErrors)}
        </div>
      ))}
    </>
  );
};

export const LaunchFormInputsImpl: React.RefForwardingComponent<
  LaunchFormInputsRef,
  LaunchFormInputsProps
> = ({ state, variant }, ref) => {
  const { parsedInputs, unsupportedRequiredInputs, showErrors } = state.context;
  const { getValues, inputs, validate } = useFormInputsState(parsedInputs);
  React.useImperativeHandle(ref, () => ({
    getValues,
    validate,
  }));

  return isEnterInputsState(state) ? (
    <section title={formStrings.inputs}>
      {state.matches(LaunchState.UNSUPPORTED_INPUTS) ? (
        <UnsupportedRequiredInputsError inputs={unsupportedRequiredInputs} variant={variant} />
      ) : (
        <RenderFormInputs inputs={inputs} showErrors={showErrors} variant={variant} />
      )}
    </section>
  ) : null;
};

/** Renders an array of `ParsedInput` values using the appropriate
 * components. A `ref` to this component is used to access the current
 * form values and trigger manual validation if needed.
 */
export const LaunchFormInputs = React.forwardRef(LaunchFormInputsImpl);
