import { ConvertFlyteDagToReactFlows } from 'components/flytegraph/ReactFlow/transformerDAGToReactFlow';
import * as React from 'react';
import { RFWrapperProps, RFGraphTypes, ConvertDagProps } from './types';
import { getRFBackground } from './utils';
import { ReactFlowWrapper } from './ReactFlowWrapper';
import { Legend } from './NodeStatusLegend';

/**
 * Renders workflow graph using React Flow.
 * @param props.data    DAG from transformerWorkflowToDAG
 * @returns ReactFlow Graph as <ReactFlowWrapper>
 */
const ReactFlowGraphComponent = props => {
    const { data, onNodeSelectionChanged, nodeExecutionsById } = props;
    const rfGraphJson = ConvertFlyteDagToReactFlows({
        root: data,
        nodeExecutionsById: nodeExecutionsById,
        onNodeSelectionChanged: onNodeSelectionChanged,
        maxRenderDepth: 1
    } as ConvertDagProps);

    const backgroundStyle = getRFBackground().nested;
    const ReactFlowProps: RFWrapperProps = {
        backgroundStyle,
        rfGraphJson: rfGraphJson,
        type: RFGraphTypes.main
    };

    const containerStyle: React.CSSProperties = {
        display: 'flex',
        flex: `1 1 100%`,
        flexDirection: 'column',
        minHeight: '100px',
        minWidth: '200px'
    };

    return (
        <div style={containerStyle}>
            <Legend />
            <ReactFlowWrapper {...ReactFlowProps} />
        </div>
    );
};

export default ReactFlowGraphComponent;
