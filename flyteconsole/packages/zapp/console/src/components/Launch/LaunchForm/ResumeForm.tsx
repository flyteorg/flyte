import { CompiledNode } from 'models/Node/types';
import * as React from 'react';
import { useState } from 'react';
import { createInputValueCache, InputValueCacheContext } from './inputValueCache';
import { ResumeSignalForm } from './ResumeSignalForm';
import { BaseLaunchFormProps, TaskInitialLaunchParameters } from './types';

interface ResumeFormProps extends BaseLaunchFormProps {
  compiledNode: CompiledNode;
  initialParameters?: TaskInitialLaunchParameters;
  nodeId: string;
}

/** Renders the form for requesting a resume request on a gate node */
export const ResumeForm: React.FC<ResumeFormProps> = (props) => {
  const [inputValueCache] = useState(createInputValueCache());

  return (
    <InputValueCacheContext.Provider value={inputValueCache}>
      <ResumeSignalForm {...props} />
    </InputValueCacheContext.Provider>
  );
};
