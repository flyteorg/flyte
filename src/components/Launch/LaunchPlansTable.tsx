import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { ListProps } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import {
    DataTable,
    DataTableRowClickEventParams,
    KeyedColumnProps
} from 'components/Tables';
import { LaunchPlan } from 'models';
import * as React from 'react';
import { TableCellDataGetterParams } from 'react-virtualized';
import { history, Routes } from 'routes';
import {
    launchPlansTableColumnWidths,
    launchPlansTableRowHeight
} from './constants';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        marginBottom: theme.spacing(2)
    }
}));

export type LaunchPlansTableProps = ListProps<LaunchPlan>
export interface CellDataGetterParams extends TableCellDataGetterParams {
    rowData: LaunchPlan;
}

export const launchPlansTableColumns: KeyedColumnProps[] = [
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) => rowData.id.name,
        dataKey: 'id',
        flexGrow: 1,
        flexShrink: 0,
        key: 'name',
        label: 'name',
        width: launchPlansTableColumnWidths.name
    },
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) =>
            rowData.id.version,
        dataKey: 'version',
        flexGrow: 1,
        flexShrink: 0,
        key: 'version',
        label: 'version',
        width: launchPlansTableColumnWidths.version
    }
];

export const LaunchPlansTable: React.FC<LaunchPlansTableProps> = props => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const onRowClick = ({ data }: DataTableRowClickEventParams<LaunchPlan>) => {
        const { project, domain, name, version } = data.spec.workflowId;
        history.push(
            // TODO: Determine which details of a LP are useful to show.
            Routes.WorkflowVersionDetails.makeUrl(
                project,
                domain,
                name,
                version!
            )
        );
    };
    const retry = () => props.fetch();
    return (
        <div className={classnames(commonStyles.flexFill, styles.container)}>
            <DataTable
                {...props}
                columns={launchPlansTableColumns}
                onRetry={retry}
                onRowClick={onRowClick}
                rowHeight={launchPlansTableRowHeight}
            />
        </div>
    );
};
