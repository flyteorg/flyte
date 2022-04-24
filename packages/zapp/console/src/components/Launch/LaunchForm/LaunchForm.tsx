import * as React from 'react';
import { createInputValueCache, InputValueCacheContext } from './inputValueCache';
import { LaunchTaskForm } from './LaunchTaskForm';
import { LaunchWorkflowForm } from './LaunchWorkflowForm';
import { LaunchFormProps, LaunchWorkflowFormProps } from './types';

function isWorkflowPropsObject(props: LaunchFormProps): props is LaunchWorkflowFormProps {
  return (props as LaunchWorkflowFormProps).workflowId !== undefined;
}

/** Renders the form for initiating a Launch request based on a Workflow or Task */
export const LaunchForm: React.FC<LaunchFormProps> = (props) => {
  const [inputValueCache] = React.useState(createInputValueCache());

  return (
    <InputValueCacheContext.Provider value={inputValueCache}>
      {isWorkflowPropsObject(props) ? (
        <LaunchWorkflowForm {...props} />
      ) : (
        <LaunchTaskForm {...props} />
      )}
    </InputValueCacheContext.Provider>
  );
};
