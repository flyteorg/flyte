import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { useTabState } from 'components/hooks/useTabState';
import { secondaryBackgroundColor } from 'components/Theme/constants';
import { Execution, NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { useState, useEffect } from 'react';
import { NodeExecutionsRequestConfigContext } from '../contexts';
import { ExecutionFilters } from '../ExecutionFilters';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
import { NodeExecutionsTable } from '../Tables/NodeExecutionsTable';
import { tabs } from './constants';
import { ExecutionChildrenLoader } from './ExecutionChildrenLoader';
import { useExecutionNodeViewsState } from './useExecutionNodeViewsState';

const useStyles = makeStyles((theme: Theme) => ({
    filters: {
        paddingLeft: theme.spacing(3)
    },
    nodesContainer: {
        borderTop: `1px solid ${theme.palette.divider}`,
        display: 'flex',
        flex: '1 1 100%',
        flexDirection: 'column',
        minHeight: 0
    },
    tabs: {
        background: secondaryBackgroundColor,
        paddingLeft: theme.spacing(3.5)
    }
}));

export interface ExecutionNodeViewsProps {
    execution: Execution;
}

/** Contains the available ways to visualize the nodes of a WorkflowExecution */
export const ExecutionNodeViews: React.FC<ExecutionNodeViewsProps> = ({
    execution
}) => {
    const styles = useStyles();
    const filterState = useNodeExecutionFiltersState();
    const tabState = useTabState(tabs, tabs.nodes.id);
    const [graphStateReady, setGraphStateReady] = useState(false);
    const {
        closure: { abortMetadata }
    } = execution;

    /* We want to maintain the filter selection when switching away from the Nodes
    tab and back, but do not want to filter the nodes when viewing the graph. So,
    we will only pass filters to the execution state when on the nodes tab. */
    const appliedFilters =
        tabState.value === tabs.nodes.id ? filterState.appliedFilters : [];

    const {
        nodeExecutionsQuery,
        nodeExecutionsRequestConfig
    } = useExecutionNodeViewsState(execution, appliedFilters);

    const renderNodeExecutionsTable = (nodeExecutions: NodeExecution[]) => (
        <NodeExecutionsRequestConfigContext.Provider
            value={nodeExecutionsRequestConfig}
        >
            <NodeExecutionsTable
                abortMetadata={abortMetadata ?? undefined}
                nodeExecutions={nodeExecutions}
            />
        </NodeExecutionsRequestConfigContext.Provider>
    );

    const renderExecutionLoader = (nodeExecutions: NodeExecution[]) => {
        return (
            <ExecutionChildrenLoader
                nodeExecutions={nodeExecutions}
                workflowId={execution.closure.workflowId}
            />
        );
    };

    return (
        <>
            <Tabs className={styles.tabs} {...tabState}>
                <Tab value={tabs.nodes.id} label={tabs.nodes.label} />
                <Tab value={tabs.graph.id} label={tabs.graph.label} />
            </Tabs>
            <div className={styles.nodesContainer}>
                {tabState.value === tabs.nodes.id && (
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
                {tabState.value === tabs.graph.id && (
                    <WaitForQuery
                        errorComponent={DataError}
                        query={nodeExecutionsQuery}
                    >
                        {renderExecutionLoader}
                    </WaitForQuery>
                )}
            </div>
        </>
    );
};
