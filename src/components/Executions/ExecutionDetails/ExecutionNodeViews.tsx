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
import { tabs } from './constants';
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
        workflow,
        nodeExecutions,
        nodeExecutionsRequestConfig
    } = useWorkflowExecutionState(execution, appliedFilters);

    return (
        <WaitForData {...workflow}>
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
                {tabState.value === tabs.graph.id && (
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
