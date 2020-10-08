import { RefObject } from 'react';
import { WorkflowLaunchContext } from './launchMachine';
import { LaunchFormInputsRef } from './types';

export async function validate(
    formInputsRef: RefObject<LaunchFormInputsRef>,
    {}: WorkflowLaunchContext
) {
    if (formInputsRef.current === null) {
        throw new Error('Unexpected empty form inputs ref');
    }

    if (!formInputsRef.current.validate()) {
        throw new Error(
            'Some inputs have errors. Please correct them before submitting.'
        );
    }
}
