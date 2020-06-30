import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import * as React from 'react';
const { useContext } = React;
import { noExecutionsFoundString } from 'common/constants';
import { NonIdealState, SectionHeader } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { TaskExecutionsList } from 'components/Executions';
import { ExecutionStatusBadge } from 'components/Executions/ExecutionStatusBadge';
import { useTabState } from 'components/hooks/useTabState';
import {
    NodeDetailsProps,
    SelectNode,
    TaskNodeDetails
} from 'components/WorkflowGraph';
import { useStyles as useBaseStyles } from 'components/WorkflowGraph/NodeDetails/styles';
import { NodeExecutionsContext } from '../../contexts';
import { NodeExecutionData } from '../NodeExecutionData';
import { NodeExecutionInputs } from '../NodeExecutionInputs';
import { NodeExecutionOutputs } from '../NodeExecutionOutputs';

const tabIds = {
    executions: 'data',
    inputs: 'inputs',
    outputs: 'outputs'
};

const NoTaskDetailsAvailable: React.FC = () => (
    <NonIdealState
        description="This node has no task associated with it"
        title="No details available"
        size="small"
    />
);

const NoExecutionsAvailable: React.FC = () => (
    <NonIdealState
        description="This node has not been executed"
        title={noExecutionsFoundString}
        size="small"
    />
);

/** DetailsPanel content which renders execution information about a given
 * graph node.
 */
export const TaskExecutionNodeDetails: React.FC<NodeDetailsProps> = props => {
    const { node } = props;
    const nodeId = node ? node.id : '';
    const nodeExecutions = useContext(NodeExecutionsContext);
    const tabState = useTabState(tabIds, tabIds.executions);
    const execution = nodeExecutions[nodeId];
    const baseStyles = useBaseStyles();
    const commonStyles = useCommonStyles();

    if (!node) {
        return <SelectNode />;
    }

    let taskContent;
    // Only supporting TaskNodes for now
    if (node.taskNode) {
        taskContent = <TaskNodeDetails taskId={node.taskNode.referenceId} />;
    } else {
        taskContent = <NoTaskDetailsAvailable />;
    }

    let statusContent;
    let executionDetailsContent;
    if (execution) {
        statusContent = (
            <ExecutionStatusBadge phase={execution.closure.phase} type="node" />
        );
        executionDetailsContent = (
            <TaskExecutionsList nodeExecution={execution} />
        );
    } else {
        executionDetailsContent = <NoExecutionsAvailable />;
    }

    return (
        <section className={baseStyles.container}>
            <header className={baseStyles.header}>
                <div className={baseStyles.headerContent}>
                    <SectionHeader title={node.id} />
                    {statusContent}
                </div>
            </header>
            <Tabs {...tabState} className={baseStyles.tabs}>
                <Tab value={tabIds.executions} label="Executions" />
                {execution && <Tab value={tabIds.inputs} label="Inputs" />}
                {execution && <Tab value={tabIds.outputs} label="Outputs" />}
            </Tabs>
            <div className={baseStyles.content}>
                {tabState.value === tabIds.executions &&
                    executionDetailsContent}
                {tabState.value === tabIds.inputs && (
                    <NodeExecutionInputs execution={execution} />
                )}
                {tabState.value === tabIds.outputs && (
                    <NodeExecutionOutputs execution={execution} />
                )}
            </div>
        </section>
    );
};
