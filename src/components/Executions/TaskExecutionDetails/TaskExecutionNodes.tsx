import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { useConditionalQuery } from 'components/hooks/useConditionalQuery';
import { useTabState } from 'components/hooks/useTabState';
import { every } from 'lodash';
import { limits } from 'models/AdminEntity/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { executionSortFields } from 'models/Execution/constants';
import { NodeExecution, TaskExecution } from 'models/Execution/types';
import * as React from 'react';
import { useQueryClient } from 'react-query';
import { executionRefreshIntervalMs } from '../constants';
import { NodeExecutionsRequestConfigContext } from '../contexts';
import { ExecutionFilters } from '../ExecutionFilters';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
import { makeTaskExecutionChildListQuery } from '../nodeExecutionQueries';
import { NodeExecutionsTable } from '../Tables/NodeExecutionsTable';
import { nodeExecutionIsTerminal, taskExecutionIsTerminal } from '../utils';

const useStyles = makeStyles((theme: Theme) => ({
    filters: {
        paddingLeft: theme.spacing(3)
    },
    nodesContainer: {
        borderTop: `1px solid ${theme.palette.divider}`,
        display: 'flex',
        flex: '1 1 100%',
        flexDirection: 'column'
    },
    tabs: {
        paddingLeft: theme.spacing(3.5)
    }
}));

interface TaskExecutionNodesProps {
    taskExecution: TaskExecution;
}

const tabIds = {
    nodes: 'nodes'
};

/** Contains the content for viewing child NodeExecutions for a TaskExecution */
export const TaskExecutionNodes: React.FC<TaskExecutionNodesProps> = ({
    taskExecution
}) => {
    const styles = useStyles();
    const filterState = useNodeExecutionFiltersState();
    const tabState = useTabState(tabIds, tabIds.nodes);

    const requestConfig = React.useMemo(
        () => ({
            filter: filterState.appliedFilters,
            limit: limits.NONE,
            sort: {
                key: executionSortFields.createdAt,
                direction: SortDirection.ASCENDING
            }
        }),
        [filterState.appliedFilters]
    );

    const shouldEnableQuery = (executions: NodeExecution[]) =>
        !every(executions, nodeExecutionIsTerminal) ||
        !taskExecutionIsTerminal(taskExecution);

    const nodeExecutionsQuery = useConditionalQuery(
        {
            ...makeTaskExecutionChildListQuery(
                useQueryClient(),
                taskExecution.id,
                requestConfig
            ),
            refetchInterval: executionRefreshIntervalMs
        },
        shouldEnableQuery
    );

    const renderNodeExecutionsTable = (nodeExecutions: NodeExecution[]) => (
        <NodeExecutionsRequestConfigContext.Provider value={requestConfig}>
            {/* TODO: legacy code - looks like it's never called
                If code still in use NodeExecutionsTable should be wrapped by 
                NodeExecutionDetailsContextProvider here or in one of the parent components.
            */}
            <NodeExecutionsTable nodeExecutions={nodeExecutions} />
        </NodeExecutionsRequestConfigContext.Provider>
    );

    return (
        <>
            <Tabs className={styles.tabs} {...tabState}>
                <Tab value={tabIds.nodes} label="Nodes" />
            </Tabs>
            <div className={styles.nodesContainer}>
                {tabState.value === tabIds.nodes && (
                    <>
                        <div className={styles.filters}>
                            <ExecutionFilters {...filterState} />
                        </div>
                        <WaitForQuery
                            errorComponent={DataError}
                            query={nodeExecutionsQuery}
                        >
                            {renderNodeExecutionsTable}
                        </WaitForQuery>
                    </>
                )}
            </div>
        </>
    );
};
