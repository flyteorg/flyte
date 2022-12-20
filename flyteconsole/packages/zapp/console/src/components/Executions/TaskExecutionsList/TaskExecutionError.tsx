import { ExpandableContentLink } from 'components/common/ExpandableContentLink';
import { useCommonStyles } from 'components/common/styles';
import { ExecutionError } from 'models/Execution/types';
import * as React from 'react';

/** Renders an expandable error for a `TaskExecution` */
export const TaskExecutionError: React.FC<{ error: ExecutionError }> = ({ error }) => {
  const commonStyles = useCommonStyles();
  const renderContent = () => <div className={commonStyles.codeBlock}>{error.message}</div>;
  return (
    <ExpandableContentLink
      allowCollapse={true}
      collapsedText="Show Error"
      expandedText="Hide Error"
      renderContent={renderContent}
    />
  );
};
