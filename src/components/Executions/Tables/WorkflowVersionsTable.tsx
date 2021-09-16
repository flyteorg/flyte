import * as classnames from 'classnames';
import { noWorkflowVersionsFoundString } from 'common/constants';
import { useCommonStyles } from 'components/common/styles';
import { ListProps } from 'components/common/types';
import { DataList, DataListRef } from 'components/Tables/DataList';
import { Workflow } from 'models/Workflow/types';
import { Identifier } from 'models/Common/types';
import * as React from 'react';
import { useParams } from 'react-router';
import { history } from 'routes/history';
import { ListRowRenderer } from 'react-virtualized';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { useExecutionTableStyles } from './styles';
import { useWorkflowExecutionsTableState } from './useWorkflowExecutionTableState';
import { useWorkflowVersionsTableColumns } from './useWorkflowVersionsTableColumns';
import { WorkflowVersionRow } from './WorkflowVersionRow';
import { Routes } from 'routes/routes';

interface WorkflowVersionsTableProps extends ListProps<Workflow> {
    versionView?: boolean;
}

interface WorkflowVersionRouteParams {
    workflowVersion: string;
}

/**
 * Renders a table of WorkflowVersion records.
 * @param props
 * @constructor
 */
export const WorkflowVersionsTable: React.FC<WorkflowVersionsTableProps> = props => {
    const { value: workflows, versionView } = props;
    const state = useWorkflowExecutionsTableState();
    const commonStyles = useCommonStyles();
    const tableStyles = useExecutionTableStyles();
    const listRef = React.useRef<DataListRef>(null);
    const { workflowVersion } = useParams<WorkflowVersionRouteParams>();

    const columns = useWorkflowVersionsTableColumns();

    const retry = () => props.fetch();

    const handleClickRow = React.useCallback(
        ({ project, name, domain, version }: Identifier) => () => {
            history.push(
                Routes.WorkflowVersionDetails.makeUrl(
                    project,
                    domain,
                    name,
                    version
                )
            );
        },
        []
    );

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
                onClick={versionView ? handleClickRow(workflow.id) : undefined}
                versionView={versionView}
                isChecked={workflowVersion === workflow.id.version}
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
            <ExecutionsTableHeader
                versionView={versionView}
                columns={columns}
            />
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
