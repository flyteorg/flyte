import * as React from 'react';

import { defaultGraphConfig } from './constants';
import { GraphProps } from './types';

import { InteractiveViewBox } from './InteractiveViewBox';
import { Layout } from './Layout';
import { RenderedGraph } from './RenderedGraph';

/** Renders a prepared DAG (an array of nodes each containing an array of parent
 *  ids) using the specified config. All config values are optional, and the
 * Graph component uses sensible defaults.
 */
export class Graph<T> extends React.Component<GraphProps<T>> {
    public render() {
        const config = { ...defaultGraphConfig, ...this.props.config };
        return (
            <Layout data={this.props.data} config={config}>
                {({ rootNode, height, width }) => (
                    <InteractiveViewBox width={width} height={height}>
                        {({ viewBox }) => (
                            <RenderedGraph
                                {...this.props}
                                config={config}
                                height={height}
                                rootNode={rootNode}
                                viewBox={viewBox}
                                width={width}
                            />
                        )}
                    </InteractiveViewBox>
                )}
            </Layout>
        );
    }
}
