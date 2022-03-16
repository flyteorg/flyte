import { DumpJSON } from 'components/common/DumpJSON';
import { useCommonStyles } from 'components/common/styles';
import { DataError } from 'components/Errors/DataError';
import { TaskTemplate } from 'models/Task/types';
import * as React from 'react';

/** Render the task template for a given NodeExecution */
export const NodeExecutionTaskDetails: React.FC<{
  taskTemplate: TaskTemplate;
}> = ({ taskTemplate }) => {
  const commonStyles = useCommonStyles();
  const content = taskTemplate ? (
    <DumpJSON value={taskTemplate} />
  ) : (
    <DataError errorTitle="No task information found" />
  );
  return (
    <div className={commonStyles.detailsPanelCard}>
      <div className={commonStyles.detailsPanelCardContent}>{content}</div>
    </div>
  );
};
