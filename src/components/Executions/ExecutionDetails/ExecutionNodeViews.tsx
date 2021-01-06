import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { useTabState } from 'components/hooks/useTabState';
import { secondaryBackgroundColor } from 'components/Theme';
import { Execution, NodeExecution } from 'models';
import * as React from 'react';
import { NodeExecutionsRequestConfigContext } from '../contexts';
import { ExecutionFilters } from '../ExecutionFilters';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
import { NodeExecutionsTable } from '../Tables/NodeExecutionsTable';
import { tabs } from './constants';
import { ExecutionWorkflowGraph } from './ExecutionWorkflowGraph';
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
            <NodeExecutionsTable nodeExecutions={nodeExecutions} />
        </NodeExecutionsRequestConfigContext.Provider>
    );

    const renderExecutionWorkflowGraph = (nodeExecutions: NodeExecution[]) => (
        <ExecutionWorkflowGraph
            nodeExecutions={nodeExecutions}
            workflowId={execution.closure.workflowId}
        />
    );

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
                        <WaitForQuery query={nodeExecutionsQuery}>
                            {renderNodeExecutionsTable}
                        </WaitForQuery>
                    </>
                )}
                {tabState.value === tabs.graph.id && (
                    <WaitForQuery query={nodeExecutionsQuery}>
                        {renderExecutionWorkflowGraph}
                    </WaitForQuery>
                )}
            </div>
        </>
    );
};
