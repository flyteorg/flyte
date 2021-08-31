import * as classnames from 'classnames';
import { noWorkflowVersionsFoundString } from 'common/constants';
import { useCommonStyles } from 'components/common/styles';
import { ListProps } from 'components/common/types';
import { DataList, DataListRef } from 'components/Tables/DataList';
import { Workflow } from 'models/Workflow/types';
import * as React from 'react';
import { ListRowRenderer } from 'react-virtualized';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { useExecutionTableStyles } from './styles';
import { useWorkflowExecutionsTableState } from './useWorkflowExecutionTableState';
import { useWorkflowVersionsTableColumns } from './useWorkflowVersionsTableColumns';
import { WorkflowVersionRow } from './WorkflowVersionRow';

/**
 * Renders a table of WorkflowVersion records.
 * @param props
 * @constructor
 */
export const WorkflowVersionsTable: React.FC<ListProps<Workflow>> = props => {
    const { value: workflows } = props;
    const state = useWorkflowExecutionsTableState();
    const commonStyles = useCommonStyles();
    const tableStyles = useExecutionTableStyles();
    const listRef = React.useRef<DataListRef>(null);

    const columns = useWorkflowVersionsTableColumns();

    const retry = () => props.fetch();

    // Custom renderer to allow us to append error content to workflow versions which
    // are in a failed state
    const rowRenderer: ListRowRenderer = rowProps => {
        const workflow = workflows[rowProps.index];
        return (
            <WorkflowVersionRow
                {...rowProps}
                columns={columns}
                workflow={workflow}
                state={state}
            />
        );
    };

    return (
        <div
            className={classnames(
                tableStyles.tableContainer,
                commonStyles.flexFill
            )}
        >
            <ExecutionsTableHeader columns={columns} />
            <DataList
                {...props}
                onRetry={retry}
                noRowsContent={noWorkflowVersionsFoundString}
                moreItemsAvailable={false}
                ref={listRef}
                rowContentRenderer={rowRenderer}
            />
        </div>
    );
};
