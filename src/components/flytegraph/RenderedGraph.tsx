import * as React from 'react';
import { Arrowhead } from './Arrowhead';
import { componentIds } from './constants';
import { DragFilteringClickHandler } from './DragAllowingClickHandler';
import { Node as DefaultNode } from './Node';
import { NodeLink as DefaultNodeLink } from './NodeLink';
import { NodeRenderer, RenderedGraphProps } from './types';

interface RenderedGraphState {
  hoveredNodes: Record<string, boolean>;
}

/** Renders a DAG that has already undergone layout into a specified height and
 * width. These values are usually created by `layoutGraph` or the `Layout`
 * component.
 */
export class RenderedGraph<T> extends React.Component<RenderedGraphProps<T>, RenderedGraphState> {
  state: RenderedGraphState = {
    hoveredNodes: {},
  };

  private containerClickHandler: DragFilteringClickHandler;

  constructor(props: RenderedGraphProps<T>) {
    super(props);
    this.containerClickHandler = new DragFilteringClickHandler(this.onContainerClick);
  }

  public render() {
    const {
      config,
      rootNode,
      height,
      linkRenderer: NodeLink = DefaultNodeLink,
      nodeRenderer = DefaultNode,
      nodeTextRender,
      selectedNodes = [],
      width,
    } = this.props;

    const viewBox = this.props.viewBox || `0 0 ${this.props.width} ${this.props.height}`;

    // This is to help the compiler with type inference
    const Node = nodeRenderer as NodeRenderer<T>;

    const { hoveredNodes } = this.state;
    const selectedNodesById = selectedNodes.reduce<Record<string, boolean>>(
      (out, nodeId) => Object.assign(out, { [nodeId]: true }),
      {},
    );
    const {
      nodeLink: { strokeColor: linkStrokeColor },
    } = config;

    const links = rootNode.links();
    const nodes = rootNode.descendants();

    return (
      <svg width="100%" height="100%" viewBox={viewBox} xmlns="http://www.w3.org/2000/svg">
        <defs>
          <Arrowhead id={componentIds.arrowhead} fill={linkStrokeColor} />
        </defs>
        {/* Content container */}
        <g>
          {/* Empty rect behind everything for background clicks */}
          <rect
            fill="none"
            height={height}
            onMouseDown={this.containerClickHandler.onMouseDown}
            onMouseUp={this.containerClickHandler.onMouseUp}
            onMouseMove={this.containerClickHandler.onMouseMove}
            pointerEvents="visible"
            stroke="none"
            width={width}
          />
          {/* Links go first, so that Nodes draw over them */}
          <g fill="none">
            {links.map((link) => (
              <NodeLink
                key={`${link.source.id}/${link.target.id}`}
                link={link}
                config={config.nodeLink}
              />
            ))}
          </g>
          {/* Node layer */}
          <g>
            {nodes.map((node) => (
              <g key={node.id} transform={`translate(${node.x}, ${node.y})`}>
                <Node
                  node={node}
                  config={config.node}
                  hovered={!!hoveredNodes[node.id]}
                  onEnter={this.onNodeEnter.bind(this, node.id)}
                  onClick={this.onNodeClick.bind(this, node.id)}
                  onLeave={this.onNodeLeave.bind(this, node.id)}
                  selected={!!selectedNodesById[node.id]}
                  textRenderer={nodeTextRender}
                />
              </g>
            ))}
          </g>
        </g>
      </svg>
    );
  }

  private clearSelection = () => {
    // Don't need to trigger a re-render if selection is already empty
    if (!this.props.selectedNodes || this.props.selectedNodes.length === 0) {
      return;
    }
    this.props.onNodeSelectionChanged && this.props.onNodeSelectionChanged([]);
  };

  private onNodeEnter = (id: string) => {
    const hoveredNodes: Dictionary<boolean> = {
      ...this.state.hoveredNodes,
      [id]: true,
    };
    this.setState({ hoveredNodes });
    this.props.onNodeEnter && this.props.onNodeEnter(id);
  };

  private onNodeLeave = (id: string) => {
    const hoveredNodes: Dictionary<boolean> = {
      ...this.state.hoveredNodes,
    };
    delete hoveredNodes[id];
    this.setState({ hoveredNodes });
    this.props.onNodeLeave && this.props.onNodeLeave(id);
  };

  private onNodeSelect = (id: string) => {
    // Only supporting single-select at the moment. So if we clicked
    // a selected node, reset it. Otherwise, the clicked node is the
    // new selection
    if (this.props.selectedNodes && this.props.selectedNodes.includes(id)) {
      return this.clearSelection();
    }

    this.props.onNodeSelectionChanged && this.props.onNodeSelectionChanged([id]);
  };

  private onNodeClick = (id: string) => {
    this.props.onNodeClick && this.props.onNodeClick(id);
    this.onNodeSelect(id);
  };

  private onContainerClick = () => {
    this.clearSelection();
  };
}
