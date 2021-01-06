import Chip from '@material-ui/core/Chip';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { getScheduleFrequencyString } from 'common/formatters';
import { ListProps } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import {
    DataTable,
    DataTableRowClickEventParams,
    KeyedColumnProps,
    TableCellProps
} from 'components/Tables';
import { LaunchPlan, Schedule } from 'models';
import * as React from 'react';
import { TableCellDataGetterParams } from 'react-virtualized';
import { history, Routes } from 'routes';
import {
    launchPlansTableRowHeight,
    schedulesTableColumnsWidths
} from './constants';
import { isLaunchPlanActive } from './utils';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        marginBottom: theme.spacing(2)
    }
}));

export type SchedulesTableProps = ListProps<LaunchPlan>
interface CellDataGetterParams extends TableCellDataGetterParams {
    rowData: LaunchPlan;
}

export const schedulesTableColumns: KeyedColumnProps[] = [
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) => rowData.id.name,
        dataKey: 'id',
        flexGrow: 2,
        flexShrink: 0,
        key: 'name',
        label: 'name',
        width: schedulesTableColumnsWidths.name
    },
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) =>
            rowData.spec.entityMetadata.schedule,
        cellRenderer: ({ cellData: schedule }: TableCellProps<Schedule>) =>
            getScheduleFrequencyString(schedule),
        dataKey: 'closure',
        flexGrow: 1,
        flexShrink: 0,
        key: 'frequency',
        label: 'frequency',
        width: schedulesTableColumnsWidths.frequency
    },
    {
        cellDataGetter: ({ rowData }: CellDataGetterParams) =>
            isLaunchPlanActive(rowData),
        cellRenderer: ({ cellData: active }: TableCellProps<boolean>) => {
            return active ? <Chip color="primary" label="active" /> : null;
        },
        dataKey: 'closure',
        flexGrow: 0,
        flexShrink: 0,
        headerRenderer: () => null,
        key: 'active',
        label: 'active',
        width: schedulesTableColumnsWidths.active
    }
];

export const SchedulesTable: React.FC<SchedulesTableProps> = props => {
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
                version
            )
        );
    };

    const retry = () => props.fetch();
    return (
        <div className={classnames(commonStyles.flexFill, styles.container)}>
            <DataTable
                {...props}
                columns={schedulesTableColumns}
                onRetry={retry}
                onRowClick={onRowClick}
                rowHeight={launchPlansTableRowHeight}
            />
        </div>
    );
};
