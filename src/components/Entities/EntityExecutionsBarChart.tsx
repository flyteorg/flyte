import Typography from '@material-ui/core/Typography';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { useWorkflowExecutions } from 'components/hooks/useWorkflowExecutions';
import { SortDirection } from 'models/AdminEntity/types';
import { ResourceIdentifier } from 'models/Common/types';
import { Execution } from 'models/Execution/types';
import { executionSortFields } from 'models/Execution/constants';
import * as React from 'react';
import { executionFilterGenerator } from './generators';
import { BarChart } from 'components/common/BarChart';
import {
    getWorkflowExecutionPhaseConstants,
    getWorkflowExecutionTimingMS
} from '../Executions/utils';
import { formatDateUTC } from 'common/formatters';
import { timestampToDate } from 'common/utils';

const useStyles = makeStyles((theme: Theme) => ({
    header: {
        paddingBottom: theme.spacing(1),
        paddingLeft: theme.spacing(1),
        borderBottom: `1px solid ${theme.palette.divider}`
    },
    body: {
        margin: theme.spacing(1)
    }
}));

export interface EntityExecutionsBarChartProps {
    id: ResourceIdentifier;
}

const getExecutionTimeData = (exectuions: Execution[], fillSize = 100) => {
    const newExecutions = exectuions.map(execution => {
        return {
            value: getWorkflowExecutionTimingMS(execution)?.duration || 1,
            color: getWorkflowExecutionPhaseConstants(execution.closure.phase)
                .badgeColor
        };
    });
    if (newExecutions.length >= fillSize) {
        return newExecutions.slice(0, fillSize);
    }
    return new Array(fillSize - newExecutions.length)
        .fill(0)
        .map(_ => ({
            value: 1,
            color: '#e5e5e5'
        }))
        .concat(newExecutions);
};

const getStartExecutionTime = (executions: Execution[]) => {
    if (executions.length === 0 || !executions[0].closure.startedAt) {
        return '';
    }
    return formatDateUTC(timestampToDate(executions[0].closure.startedAt));
};

/**
 * The tab/page content for viewing a workflow's executions as bar chart
 * @param id
 * @constructor
 */
export const EntityExecutionsBarChart: React.FC<EntityExecutionsBarChartProps> = ({
    id
}) => {
    const styles = useStyles();
    const { domain, project, resourceType } = id;
    const filtersState = useWorkflowExecutionFiltersState();
    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.ASCENDING
    };

    const baseFilters = React.useMemo(
        () => executionFilterGenerator[resourceType](id),
        [id]
    );

    const executions = useWorkflowExecutions(
        { domain, project },
        {
            sort,
            filter: [...baseFilters, ...filtersState.appliedFilters]
        }
    );

    /** Don't render component until finish fetching user profile */
    if (filtersState.filters[4].status !== 'LOADED') {
        return null;
    }

    return (
        <WaitForData {...executions}>
            <Typography className={styles.header} variant="h6">
                All Executions in the Workflow
            </Typography>
            <div className={styles.body}>
                <BarChart
                    data={getExecutionTimeData(executions.value)}
                    startDate={getStartExecutionTime(executions.value)}
                />
            </div>
        </WaitForData>
    );
};
