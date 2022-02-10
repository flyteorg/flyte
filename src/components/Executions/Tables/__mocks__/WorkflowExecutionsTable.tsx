import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { ExecutionsTableHeader } from '../ExecutionsTableHeader';
import { useExecutionTableStyles } from '../styles';
import { useWorkflowExecutionsTableColumns } from '../WorkflowExecutionTable/useWorkflowExecutionsTableColumns';
import { useWorkflowExecutionsTableState } from '../useWorkflowExecutionTableState';
import { WorkflowExecutionRow } from '../WorkflowExecutionTable/WorkflowExecutionRow';
import { WorkflowExecutionsTableProps } from '../WorkflowExecutionsTable';

/** Mocked, simpler version of WorkflowExecutionsTable which does not use a DataList since
 * that will not work in a test environment.
 */
export const WorkflowExecutionsTable: React.FC<WorkflowExecutionsTableProps> = props => {
    const { value: executions, showWorkflowName = false } = props;
    const state = useWorkflowExecutionsTableState();
    const commonStyles = useCommonStyles();
    const tableStyles = useExecutionTableStyles();
    const columns = useWorkflowExecutionsTableColumns({});

    return (
        <div
            className={classnames(
                tableStyles.tableContainer,
                commonStyles.flexFill
            )}
        >
            <ExecutionsTableHeader columns={columns} />
            {executions.map(execution => (
                <WorkflowExecutionRow
                    showWorkflowName={showWorkflowName}
                    key={execution.id.name}
                    execution={execution}
                    state={state}
                />
            ))}
        </div>
    );
};
