import { NonIdealState, SectionHeader } from 'components/common';
import { CompiledNode } from 'models';
import * as React from 'react';
import { SelectNode } from './SelectNode';
import { useStyles } from './styles';
import { TaskNodeDetails } from './TaskNodeDetails';

export interface NodeDetailsProps {
    node?: CompiledNode;
}

const NoDetailsAvailable: React.FC = () => (
    <NonIdealState
        description="This node has no task associated with it"
        size="small"
        title="No details available"
    />
);

/** DetailsPanel content which renders information about a given node in a
 * workflow graph.
 */
export const NodeDetails: React.FC<NodeDetailsProps> = props => {
    const styles = useStyles();
    const { node } = props;
    if (!node) {
        return <SelectNode />;
    }

    let content;
    let subtitle = null;
    // Only supporting TaskNodes for now
    if (node.taskNode) {
        content = <TaskNodeDetails taskId={node.taskNode.referenceId} />;
        subtitle = 'task node';
    } else {
        content = <NoDetailsAvailable />;
    }

    return (
        <section className={styles.container}>
            <SectionHeader title={node.id} subtitle={subtitle} />
            <div className={styles.content}>{content}</div>
        </section>
    );
};
