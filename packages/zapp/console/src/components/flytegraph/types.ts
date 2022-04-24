import * as React from 'react';
import { DierarchyPointNode } from 'd3-dag';

/** Graph-related types */
export interface GraphLayoutResult<T> {
  height: number;
  rootNode: RenderableNode<T>;
  width: number;
}

export interface Point {
  x: number;
  y: number;
}

export interface RenderableNode<T> extends DierarchyPointNode<T> {
  id: string;
  textWidth: number;
  x: number;
  y: number;
  links(): RenderableNodeLink<T>[];
}

export interface RenderableNodeLink<T> {
  source: RenderableNode<T>;
  target: RenderableNode<T>;
  data: {
    points: Point[];
  };
}

export interface GraphInputNode {
  id: string;
  parentIds?: string[];
}

export interface NodeConfig {
  cornerRounding: number | string;
  fillColor: string;
  fontSize: number;
  selectStrokeColor: string;
  selectStrokeWidth: number;
  strokeColor: string;
  strokeWidth: number;
  // Text color is optional, default component will choose an appropriate
  // color if not specified
  textColor?: string;
  textPadding: number;
}

export interface NodeLinkConfig {
  strokeColor: string;
  strokeWidth: number;
}

export interface GraphConfig {
  node: NodeConfig;
  nodeLink: NodeLinkConfig;
}

interface BaseNodeProps<T> {
  node: RenderableNode<T>;
  config: NodeConfig;
  selected?: boolean;
  hovered?: boolean;
  onClick?: React.MouseEventHandler<SVGElement>;
  onEnter?: React.MouseEventHandler<SVGElement>;
  onLeave?: React.MouseEventHandler<SVGElement>;
}

export interface NodeRendererProps<T> extends BaseNodeProps<T> {
  textRenderer?: NodeTextRenderer<T>;
}

export type NodeTextRendererProps<T> = BaseNodeProps<T>;

export interface NodeLinkRendererProps<T> {
  link: RenderableNodeLink<T>;
  config: NodeLinkConfig;
}

export type NodeRenderer<T> = React.ComponentType<NodeRendererProps<T>>;
export type NodeTextRenderer<T> = React.ComponentType<NodeTextRendererProps<T>>;
export type NodeLinkRenderer<T> = React.ComponentType<NodeLinkRendererProps<T>>;
export type NodeEventHandler = (id: string) => void;

export interface LayoutProps<T> {
  children: (props: GraphLayoutResult<T>) => React.ReactNode;
  data: (T & GraphInputNode)[];
  config: GraphConfig;
}

export interface GraphProps<T> {
  data: (T & GraphInputNode)[];
  config?: Partial<GraphConfig>;
  linkRenderer?: NodeLinkRenderer<T>;
  nodeRenderer?: NodeRenderer<T>;
  nodeTextRender?: NodeTextRenderer<T>;
  onNodeClick?: NodeEventHandler;
  onNodeEnter?: NodeEventHandler;
  onNodeLeave?: NodeEventHandler;
  onNodeSelectionChanged?: (selectedNodes: string[]) => void;
  selectedNodes?: string[];
}

export interface RenderedGraphProps<T> extends GraphProps<T> {
  config: GraphConfig;
  rootNode: RenderableNode<T>;
  height: number;
  width: number;
  viewBox?: string;
}
