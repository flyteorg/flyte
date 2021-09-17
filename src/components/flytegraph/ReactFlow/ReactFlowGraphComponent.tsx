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
        maxRenderDepth: 2
    } as ConvertDagProps);

    const backgroundStyle = getRFBackground().nested;
    const ReactFlowProps: RFWrapperProps = {
        backgroundStyle,
        rfGraphJson: rfGraphJson,
        type: RFGraphTypes.main
    };
    return (
        <>
            <Legend />
            <ReactFlowWrapper {...ReactFlowProps} />
        </>
    );
};

export default ReactFlowGraphComponent;
