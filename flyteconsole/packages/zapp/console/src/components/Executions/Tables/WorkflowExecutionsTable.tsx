import classnames from 'classnames';
import * as React from 'react';
import { ListRowRenderer } from 'react-virtualized';
import { noExecutionsFoundString } from 'common/constants';
import { getCacheKey } from 'components/Cache/utils';
import { useCommonStyles } from 'components/common/styles';
import { ListProps } from 'components/common/types';
import { DataList, DataListRef } from 'components/Tables/DataList';
import { Execution } from 'models/Execution/types';
import { ExecutionInputsOutputsModal } from '../ExecutionInputsOutputsModal';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { useExecutionTableStyles } from './styles';
import { useWorkflowExecutionsTableColumns } from './WorkflowExecutionTable/useWorkflowExecutionsTableColumns';
import { useWorkflowExecutionsTableState } from './useWorkflowExecutionTableState';
import { WorkflowExecutionRow } from './WorkflowExecutionTable/WorkflowExecutionRow';

export interface WorkflowExecutionsTableProps extends ListProps<Execution> {
  showWorkflowName?: boolean;
}

/** Renders a table of WorkflowExecution records. Executions with errors will
 * have an expanadable container rendered as part of the table row.
 */
export const WorkflowExecutionsTable: React.FC<WorkflowExecutionsTableProps> = (props) => {
  const { value: executions, showWorkflowName = false } = props;
  const [expandedErrors, setExpandedErrors] = React.useState<Dictionary<boolean>>({});
  const state = useWorkflowExecutionsTableState();
  const commonStyles = useCommonStyles();
  const tableStyles = useExecutionTableStyles();
  const listRef = React.useRef<DataListRef>(null);

  // Reset error expansion states whenever list changes
  React.useLayoutEffect(() => {
    setExpandedErrors({});
  }, [executions]);

  // passing an empty property list, as we only use it for table headers info here
  const columns = useWorkflowExecutionsTableColumns({});

  const retry = () => props.fetch();
  const onCloseIOModal = () => state.setSelectedIOExecution(null);
  const recomputeRow = (rowIndex: number) => {
    if (listRef.current !== null) {
      listRef.current.recomputeRowHeights(rowIndex);
    }
  };

  // Custom renderer to allow us to append error content to executions which
  // are in a failed state
  const rowRenderer: ListRowRenderer = (rowProps) => {
    const execution = executions[rowProps.index];
    const cacheKey = getCacheKey(execution.id);
    const onExpandCollapseError = (expanded: boolean) => {
      setExpandedErrors((currentExpandedErrors) => ({
        ...currentExpandedErrors,
        [cacheKey]: expanded,
      }));
      recomputeRow(rowProps.index);
    };
    return (
      <WorkflowExecutionRow
        {...rowProps}
        showWorkflowName={showWorkflowName}
        execution={execution}
        errorExpanded={!!expandedErrors[cacheKey]}
        onExpandCollapseError={onExpandCollapseError}
        state={state}
      />
    );
  };

  return (
    <div className={classnames(tableStyles.tableContainer, commonStyles.flexFill)}>
      <ExecutionsTableHeader columns={columns} />
      <DataList
        {...props}
        onRetry={retry}
        noRowsContent={noExecutionsFoundString}
        ref={listRef}
        rowContentRenderer={rowRenderer}
      />
      <ExecutionInputsOutputsModal execution={state.selectedIOExecution} onClose={onCloseIOModal} />
    </div>
  );
};
