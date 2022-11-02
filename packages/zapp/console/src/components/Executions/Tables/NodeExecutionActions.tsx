import { IconButton, Tooltip } from '@material-ui/core';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import InputsAndOutputsIcon from '@material-ui/icons/Tv';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import { RerunIcon } from '@flyteconsole/ui-atoms';
import { Identifier, ResourceIdentifier } from 'models/Common/types';
import { LaunchFormDialog } from 'components/Launch/LaunchForm/LaunchFormDialog';
import { getTask } from 'models/Task/api';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { TaskInitialLaunchParameters } from 'components/Launch/LaunchForm/types';
import { literalsToLiteralValueMap } from 'components/Launch/LaunchForm/utils';
import { useContext, useEffect, useState } from 'react';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionDetails } from '../types';
import t from './strings';
import { getNodeFrontendPhase, isNodeGateNode } from '../utils';
import { DetailsPanelContext } from '../ExecutionDetails/DetailsPanelContext';

interface NodeExecutionActionsProps {
  execution: NodeExecution;
}

export const NodeExecutionActions = ({ execution }: NodeExecutionActionsProps): JSX.Element => {
  const { compiledWorkflowClosure, getNodeExecutionDetails } = useNodeExecutionContext();
  const { setSelectedExecution } = useContext(DetailsPanelContext);

  const [showLaunchForm, setShowLaunchForm] = useState<boolean>(false);
  const [showResumeForm, setShowResumeForm] = useState<boolean>(false);
  const [nodeExecutionDetails, setNodeExecutionDetails] = useState<
    NodeExecutionDetails | undefined
  >(undefined);
  const [initialParameters, setInitialParameters] = useState<
    TaskInitialLaunchParameters | undefined
  >(undefined);

  const executionData = useNodeExecutionData(execution.id);
  const id = nodeExecutionDetails?.taskTemplate?.id;

  const isGateNode = isNodeGateNode(
    compiledWorkflowClosure?.primary.template.nodes ?? [],
    execution.id,
  );
  const phase = getNodeFrontendPhase(execution.closure.phase, isGateNode);
  const compiledNode = (compiledWorkflowClosure?.primary.template.nodes ?? []).find(
    (node) => node.id === execution.id.nodeId,
  );

  useEffect(() => {
    let isCurrent = true;
    getNodeExecutionDetails(execution).then((res) => {
      if (isCurrent) {
        setNodeExecutionDetails(res);
      }
    });
    return () => {
      isCurrent = false;
    };
  });

  useEffect(() => {
    if (!id) {
      return;
    }

    (async () => {
      const task = await getTask(id as Identifier);

      const literals = executionData.value.fullInputs?.literals;
      const taskInputsTypes = task.closure.compiledTask.template?.interface?.inputs?.variables;

      const tempInitialParameters: TaskInitialLaunchParameters = {
        values: literals && taskInputsTypes && literalsToLiteralValueMap(literals, taskInputsTypes),
        taskId: id as Identifier | undefined,
      };

      setInitialParameters(tempInitialParameters);
    })();
  }, [id]);

  // open the side panel for selected execution's detail
  const inputsAndOutputsIconOnClick = (e: React.MouseEvent<HTMLElement>) => {
    // prevent the parent row body onClick event trigger
    e.stopPropagation();
    // use null in case if there is no execution provided - when it is null will close panel
    setSelectedExecution(execution?.id ?? null);
  };

  const rerunIconOnClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setShowLaunchForm(true);
  };

  const onResumeClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setShowResumeForm(true);
  };

  return (
    <div>
      {phase === NodeExecutionPhase.PAUSED && (
        <Tooltip title={t('resumeTooltip')}>
          <IconButton onClick={onResumeClick}>
            <PlayCircleOutlineIcon />
          </IconButton>
        </Tooltip>
      )}
      <Tooltip title={t('inputsAndOutputsTooltip')}>
        <IconButton onClick={inputsAndOutputsIconOnClick}>
          <InputsAndOutputsIcon />
        </IconButton>
      </Tooltip>
      {id && initialParameters ? (
        <>
          <Tooltip title={t('rerunTooltip')}>
            <IconButton onClick={rerunIconOnClick}>
              <RerunIcon />
            </IconButton>
          </Tooltip>
          <LaunchFormDialog
            id={id as ResourceIdentifier}
            initialParameters={initialParameters}
            showLaunchForm={showLaunchForm}
            setShowLaunchForm={setShowLaunchForm}
          />
        </>
      ) : null}
      {compiledNode && (
        <LaunchFormDialog
          compiledNode={compiledNode}
          initialParameters={initialParameters}
          nodeId={execution.id.nodeId}
          showLaunchForm={showResumeForm}
          setShowLaunchForm={setShowResumeForm}
        />
      )}
    </div>
  );
};
