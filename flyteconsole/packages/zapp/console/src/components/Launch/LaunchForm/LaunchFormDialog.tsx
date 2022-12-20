import { Dialog } from '@material-ui/core';
import * as React from 'react';
import { LaunchForm } from 'components/Launch/LaunchForm/LaunchForm';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import {
  TaskInitialLaunchParameters,
  WorkflowInitialLaunchParameters,
} from 'components/Launch/LaunchForm/types';
import { CompiledNode } from 'models/Node/types';
import { ResumeForm } from './ResumeForm';

interface LaunchFormDialogProps {
  id?: ResourceIdentifier;
  initialParameters?: TaskInitialLaunchParameters | WorkflowInitialLaunchParameters;
  showLaunchForm: boolean;
  setShowLaunchForm: React.Dispatch<React.SetStateAction<boolean>>;
  compiledNode?: CompiledNode;
  nodeId?: string;
}

function getLaunchProps(id: ResourceIdentifier) {
  if (id.resourceType === ResourceType.TASK) {
    return { taskId: id };
  } else if (id.resourceType === ResourceType.WORKFLOW) {
    return { workflowId: id };
  }
  throw new Error('Unknown Resource Type');
}

export const LaunchFormDialog = ({
  id,
  initialParameters,
  showLaunchForm,
  setShowLaunchForm,
  compiledNode,
  nodeId,
}: LaunchFormDialogProps): JSX.Element => {
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
      {id ? (
        <LaunchForm
          initialParameters={initialParameters}
          onClose={onCancelLaunch}
          {...getLaunchProps(id)}
        />
      ) : compiledNode && nodeId ? (
        <ResumeForm
          initialParameters={initialParameters}
          nodeId={nodeId}
          onClose={onCancelLaunch}
          compiledNode={compiledNode}
        />
      ) : null}
    </Dialog>
  );
};
