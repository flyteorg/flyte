import { IconButton, Tooltip } from '@material-ui/core';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import InputsAndOutputsIcon from '@material-ui/icons/Tv';
import { RerunIcon } from '@flyteconsole/ui-atoms';
import { Identifier, ResourceIdentifier, Variable } from 'models/Common/types';
import { LaunchFormDialog } from 'components/Launch/LaunchForm/LaunchFormDialog';
import { getTask } from 'models/Task/api';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { TaskInitialLaunchParameters } from 'components/Launch/LaunchForm/types';
import { literalsToLiteralValueMap } from 'components/Launch/LaunchForm/utils';
import { NodeExecutionsTableState } from './types';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionDetails } from '../types';
import t from './strings';

interface NodeExecutionActionsProps {
  execution: NodeExecution;
  state: NodeExecutionsTableState;
}

export const NodeExecutionActions = (props: NodeExecutionActionsProps): JSX.Element => {
  const { execution, state } = props;

  const detailsContext = useNodeExecutionContext();
  const [showLaunchForm, setShowLaunchForm] = React.useState<boolean>(false);
  const [nodeExecutionDetails, setNodeExecutionDetails] = React.useState<
    NodeExecutionDetails | undefined
  >();
  const [taskInputsTypes, setTaskInputsTypes] = React.useState<
    Record<string, Variable> | undefined
  >();

  const executionData = useNodeExecutionData(execution.id);
  const literals = executionData.value.fullInputs?.literals;
  const id = nodeExecutionDetails?.taskTemplate?.id as ResourceIdentifier;

  React.useEffect(() => {
    detailsContext.getNodeExecutionDetails(execution).then((res) => {
      setNodeExecutionDetails(res);
    });
  });

  React.useEffect(() => {
    const fetchTask = async () => {
      const task = await getTask(id as Identifier);
      setTaskInputsTypes(task.closure.compiledTask.template?.interface?.inputs?.variables);
    };
    if (id) fetchTask();
  }, [id]);

  // open the side panel for selected execution's detail
  const inputsAndOutputsIconOnClick = (e: React.MouseEvent<HTMLElement>) => {
    // prevent the parent row body onClick event trigger
    e.stopPropagation();
    // use null in case if there is no execution provided - when it is null will close panel
    state.setSelectedExecution(execution?.id ?? null);
  };

  const rerunIconOnClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setShowLaunchForm(true);
  };

  const renderRerunAction = () => {
    if (!id) {
      return <></>;
    }

    const initialParameters: TaskInitialLaunchParameters = {
      values: literals && taskInputsTypes && literalsToLiteralValueMap(literals, taskInputsTypes),
      taskId: id as Identifier | undefined,
    };
    return (
      <>
        <Tooltip title={t('rerunTooltip')}>
          <IconButton onClick={rerunIconOnClick}>
            <RerunIcon />
          </IconButton>
        </Tooltip>
        <LaunchFormDialog
          id={id}
          initialParameters={initialParameters}
          showLaunchForm={showLaunchForm}
          setShowLaunchForm={setShowLaunchForm}
        />
      </>
    );
  };

  return (
    <div>
      <Tooltip title={t('inputsAndOutputsTooltip')}>
        <IconButton onClick={inputsAndOutputsIconOnClick}>
          <InputsAndOutputsIcon />
        </IconButton>
      </Tooltip>
      {renderRerunAction()}
    </div>
  );
};
