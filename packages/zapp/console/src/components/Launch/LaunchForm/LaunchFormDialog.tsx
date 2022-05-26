import { Dialog } from '@material-ui/core';
import * as React from 'react';
import { LaunchForm } from 'components/Launch/LaunchForm/LaunchForm';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import {
  TaskInitialLaunchParameters,
  WorkflowInitialLaunchParameters,
} from 'components/Launch/LaunchForm/types';

interface LaunchFormDialogProps {
  id: ResourceIdentifier;
  initialParameters: TaskInitialLaunchParameters | WorkflowInitialLaunchParameters;
  showLaunchForm: boolean;
  setShowLaunchForm: React.Dispatch<React.SetStateAction<boolean>>;
}

function getLaunchProps(id: ResourceIdentifier) {
  if (id.resourceType === ResourceType.TASK) {
    return { taskId: id };
  } else if (id.resourceType === ResourceType.WORKFLOW) {
    return { workflowId: id };
  }
  throw new Error('Unknown Resource Type');
}

export const LaunchFormDialog = (props: LaunchFormDialogProps): JSX.Element => {
  const { id, initialParameters, showLaunchForm, setShowLaunchForm } = props;

  const onCancelLaunch = () => setShowLaunchForm(false);

  // prevent child onclick event in the dialog triggers parent onclick event
  const dialogOnClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
  };

  return (
    <Dialog
      scroll="paper"
      maxWidth="sm"
      fullWidth={true}
      open={showLaunchForm}
      onClick={dialogOnClick}
    >
      <LaunchForm
        initialParameters={initialParameters}
        onClose={onCancelLaunch}
        {...getLaunchProps(id)}
      />
    </Dialog>
  );
};
