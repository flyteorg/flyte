import { Button } from '@material-ui/core';
import * as React from 'react';
import { ResourceIdentifier, Identifier, Variable } from 'models/Common/types';
import { getTask } from 'models/Task/api';
import { LaunchFormDialog } from 'components/Launch/LaunchForm/LaunchFormDialog';
import { NodeExecutionIdentifier } from 'models/Execution/types';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { literalsToLiteralValueMap } from 'components/Launch/LaunchForm/utils';
import { TaskInitialLaunchParameters } from 'components/Launch/LaunchForm/types';
import { NodeExecutionDetails } from '../types';
import t from './strings';

interface ExecutionDetailsActionsProps {
  className?: string;
  details: NodeExecutionDetails;
  nodeExecutionId: NodeExecutionIdentifier;
}

export const ExecutionDetailsActions = (props: ExecutionDetailsActionsProps): JSX.Element => {
  const { className, details, nodeExecutionId } = props;

  const [showLaunchForm, setShowLaunchForm] = React.useState<boolean>(false);
  const [taskInputsTypes, setTaskInputsTypes] = React.useState<
    Record<string, Variable> | undefined
  >();

  const executionData = useNodeExecutionData(nodeExecutionId);

  const id = details.taskTemplate?.id as ResourceIdentifier | undefined;

  React.useEffect(() => {
    const fetchTask = async () => {
      const task = await getTask(id as Identifier);
      setTaskInputsTypes(task.closure.compiledTask.template?.interface?.inputs?.variables);
    };
    if (id) fetchTask();
  }, [id]);

  if (!id) {
    return <></>;
  }

  const literals = executionData.value.fullInputs?.literals;

  const initialParameters: TaskInitialLaunchParameters = {
    values: literals && taskInputsTypes && literalsToLiteralValueMap(literals, taskInputsTypes),
    taskId: id as Identifier | undefined,
  };

  const rerunOnClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setShowLaunchForm(true);
  };

  return (
    <>
      <div className={className}>
        <Button variant="outlined" color="primary" onClick={rerunOnClick}>
          {t('rerun')}
        </Button>
      </div>
      <LaunchFormDialog
        id={id}
        initialParameters={initialParameters}
        showLaunchForm={showLaunchForm}
        setShowLaunchForm={setShowLaunchForm}
      />
    </>
  );
};
