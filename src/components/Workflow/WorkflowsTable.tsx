import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { ListProps } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import {
    DataTable,
    DataTableRowClickEventParams,
    KeyedColumnProps
} from 'components/Tables';
import { Workflow } from 'models';
import * as React from 'react';
import { TableCellDataGetterParams, TableCellProps } from 'react-virtualized';
import { history, Routes } from 'routes';
import {
    workflowsTableColumnWidths,
    workflowsTableRowHeight
} from './constants';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        marginBottom: theme.spacing(2)
    }
}));

export type WorkflowsTableProps = ListProps<Workflow>
export interface CellDataGetterParams extends TableCellDataGetterParams {
    rowData: Workflow;
}

export interface CellProps<T> extends TableCellProps {
    cellData?: T;
}

export const workflowTableColumns: KeyedColumnProps[] = [
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) => rowData.id.name,
        dataKey: 'id',
        flexGrow: 1,
        flexShrink: 0,
        key: 'name',
        label: 'name',
        width: workflowsTableColumnWidths.name
    },
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) =>
            rowData.id.version,
        dataKey: 'version',
        flexGrow: 1,
        flexShrink: 0,
        key: 'version',
        label: 'version',
        width: workflowsTableColumnWidths.version
    }
];

export const WorkflowsTable: React.FC<WorkflowsTableProps> = props => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();

    const onRowClick = ({ data }: DataTableRowClickEventParams<Workflow>) => {
        const {
            id: { project, domain, name, version }
        } = data;
        history.push(
            Routes.WorkflowVersionDetails.makeUrl(
                project,
                domain,
                name,
                version!
            )
        );
    };

    const retry = () => props.fetch;
    return (
        <div className={classnames(commonStyles.flexFill, styles.container)}>
            <DataTable
                {...props}
                columns={workflowTableColumns}
                onRetry={retry}
                onRowClick={onRowClick}
                rowHeight={workflowsTableRowHeight}
            />
        </div>
    );
};
