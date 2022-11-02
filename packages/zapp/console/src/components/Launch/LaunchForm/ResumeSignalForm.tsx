import { DialogContent, Typography } from '@material-ui/core';
import { getCacheKey } from 'components/Cache/utils';
import * as React from 'react';
import { useState, useContext, useEffect, useMemo } from 'react';
import { NodeExecution } from 'models/Execution/types';
import { NodeExecutionsByIdContext } from 'components/Executions/contexts';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { LiteralMapViewer } from 'components/Literals/LiteralMapViewer';
import { WaitForData } from 'components/common/WaitForData';
import t from 'components/common/strings';
import { CompiledNode } from 'models/Node/types';
import { useStyles } from './styles';
import {
  BaseInterpretedLaunchState,
  BaseLaunchFormProps,
  BaseLaunchService,
  TaskInitialLaunchParameters,
} from './types';
import { useResumeFormState } from './useResumeFormState';
import { LaunchFormInputs } from './LaunchFormInputs';
import { LaunchFormHeader } from './LaunchFormHeader';
import launchFormStrings from './strings';
import { LaunchFormActions } from './LaunchFormActions';

export interface ResumeSignalFormProps extends BaseLaunchFormProps {
  compiledNode: CompiledNode;
  initialParameters?: TaskInitialLaunchParameters;
  nodeId: string;
}

/** Renders the form for requesting a resume request on a gate node */
export const ResumeSignalForm: React.FC<ResumeSignalFormProps> = ({
  compiledNode,
  nodeId,
  onClose,
}) => {
  const { formInputsRef, state, service } = useResumeFormState({ compiledNode, nodeId, onClose });
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);
  const [nodeExecution, setNodeExecution] = useState<NodeExecution>(nodeExecutionsById[nodeId]);
  const styles = useStyles();
  const baseState = state as BaseInterpretedLaunchState;
  const baseService = service as BaseLaunchService;
  const [isError, setIsError] = useState<boolean>(false);
  const executionData = useNodeExecutionData(nodeExecution.id);

  // Any time the inputs change (even if it's just re-ordering), we must
  // change the form key so that the inputs component will re-mount.
  const formKey = useMemo<string>(() => {
    return getCacheKey(state.context.parsedInputs);
  }, [state.context.parsedInputs]);

  useEffect(() => {
    const newNodeExecution = nodeExecutionsById[nodeId];
    setNodeExecution(newNodeExecution);
  }, [nodeId]);

  return (
    <>
      <LaunchFormHeader
        title={nodeExecution.id.nodeId}
        formTitle={launchFormStrings('resumeTitle')}
      />
      <DialogContent dividers={true} className={styles.inputsSection}>
        <LaunchFormInputs
          key={formKey}
          ref={formInputsRef}
          state={baseState}
          variant="task"
          setIsError={setIsError}
        />
        <Typography variant="h6">{t('gateNodeInput')}</Typography>
        <WaitForData {...executionData}>
          <LiteralMapViewer map={executionData.value.fullInputs} />
        </WaitForData>
      </DialogContent>
      <LaunchFormActions
        state={baseState}
        service={baseService}
        onClose={onClose}
        isError={isError}
        submitTitle={launchFormStrings('resume')}
      />
    </>
  );
};
