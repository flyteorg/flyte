import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common';
import { useTabState } from 'components/hooks/useTabState';
import { secondaryBackgroundColor } from 'components/Theme';
import { Execution } from 'models';
import * as React from 'react';
import { NodeExecutionsRequestConfigContext } from '../contexts';
import { ExecutionFilters } from '../ExecutionFilters';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
import { NodeExecutionsTable } from '../Tables/NodeExecutionsTable';
import { useWorkflowExecutionState } from '../useWorkflowExecutionState';
import { ExecutionWorkflowGraph } from './ExecutionWorkflowGraph';

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

interface ExecutionNodeViewsProps {
    execution: Execution;
}

const tabIds = {
    nodes: 'nodes',
    graph: 'graph'
};

/** Contains the available ways to visualize the nodes of a WorkflowExecution */
export const ExecutionNodeViews: React.FC<ExecutionNodeViewsProps> = ({
    execution
}) => {
    const styles = useStyles();
    const filterState = useNodeExecutionFiltersState();
    const tabState = useTabState(tabIds, tabIds.nodes);

    const {
        workflow,
        nodeExecutions,
        nodeExecutionsRequestConfig
    } = useWorkflowExecutionState(execution, filterState.appliedFilters);

    return (
        <WaitForData {...workflow}>
            <Tabs className={styles.tabs} {...tabState}>
                <Tab value={tabIds.nodes} label="Nodes" />
                <Tab value={tabIds.graph} label="Graph" />
            </Tabs>
            <div className={styles.nodesContainer}>
                {tabState.value === tabIds.nodes && (
                    <>
                        <div className={styles.filters}>
                            <ExecutionFilters {...filterState} />
                        </div>
                        <WaitForData {...nodeExecutions}>
                            <NodeExecutionsRequestConfigContext.Provider
                                value={nodeExecutionsRequestConfig}
                            >
                                <NodeExecutionsTable
                                    {...nodeExecutions}
                                    moreItemsAvailable={false}
                                />
                            </NodeExecutionsRequestConfigContext.Provider>
                        </WaitForData>
                    </>
                )}
                {tabState.value === tabIds.graph && (
                    <WaitForData {...nodeExecutions}>
                        <ExecutionWorkflowGraph
                            nodeExecutions={nodeExecutions.value}
                            workflow={workflow.value}
                        />
                    </WaitForData>
                )}
            </div>
        </WaitForData>
    );
};
