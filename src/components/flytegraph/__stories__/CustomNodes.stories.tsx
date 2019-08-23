import * as React from 'react';

import { action as storybookAction } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { defaultGraphConfig } from '../constants';
import { Graph } from '../Graph';
import { measureText } from '../layoutUtils';
import { Node } from '../Node';
import { NodeText } from '../NodeText';
import { colors } from '../theme';
import { NodeRenderer, NodeRendererProps, RenderableNode } from '../types';

import customNodeData from './customNodeData';

const ColoredNode: NodeRenderer<any> = props => {
    const { node, config } = props;
    const { fill: fillColor } = node.data;
    return <Node {...props} config={{ ...config, fillColor }} />;
};

const NodeColors: React.FC<{}> = () => {
    const config = {
        ...defaultGraphConfig,
        graphScale: 1
    };
    const width = 200;
    const nodeHeight = 50;
    const padding = 20;
    const height = Object.keys(colors).length * nodeHeight + padding * 2;
    const fontSize = 14;
    return (
        <svg xmlns="http://www.w3.org/2000/svg" width={width} height={height}>
            {Object.keys(colors).map((key, idx) => {
                const color = (colors as Dictionary<string>)[key];
                const props: NodeRendererProps<any> = {
                    config: {
                        ...config.node,
                        fillColor: color
                    },
                    node: {
                        x: 0,
                        y: 0,
                        textWidth: measureText(fontSize, key).width,
                        id: key
                    } as RenderableNode<any>
                };
                return (
                    <g
                        key={key}
                        transform={`translate(100, ${idx * nodeHeight +
                            padding})`}
                    >
                        <Node {...props} />
                        <NodeText {...props} />
                    </g>
                );
            })}
        </svg>
    );
};

const stories = storiesOf('flytegraph/Graph', module);
stories.addDecorator(story => (
    <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}>
        {story()}
    </div>
));
const eventHandlers = {
    onNodeClick: storybookAction('onNodeClick'),
    onNodeEnter: storybookAction('onNodeEnter'),
    onNodeLeave: storybookAction('onNodeLeave'),
    onNodeSelectionChanged: storybookAction('onNodeSelectionChanged')
};

stories.add('custom nodes', () => (
    <Graph
        {...eventHandlers}
        nodeRenderer={ColoredNode}
        data={customNodeData}
    />
));
stories.add('node colors', () => <NodeColors />);
