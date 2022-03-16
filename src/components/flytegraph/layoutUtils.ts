import * as d3 from 'd3-dag';
import { cloneDeep } from 'lodash';
import { layoutSize, nodeSpacingMultiplier } from './constants';
import { createTimer } from './timer';
import { GraphConfig, GraphInputNode, GraphLayoutResult, RenderableNode } from './types';
import { createDebugLogger, groupBy } from './utils';

const log = createDebugLogger('layout');

interface ColumnMeasurement {
  x: number;
  maxWidth: number;
}

function findColumnMeasurements<T>(nodes: RenderableNode<T>[]): ColumnMeasurement[] {
  // sorting will modify in place, so make a copy
  const cloned = [...nodes];
  const buckets = groupBy('x', cloned);
  return buckets.map(({ x, points }) => {
    const maxWidth = points.reduce<number>((out, node) => {
      return Math.max(out, node.textWidth);
    }, 0);
    return { x, maxWidth };
  });
}

function findNeededHorizontalColumnScale(columns: ColumnMeasurement[]) {
  return columns.reduce<number>((currentMax, column, idx) => {
    if (idx === 0) {
      return currentMax;
    }
    const prevColumn = columns[idx - 1];
    const prevRightBound = prevColumn.x + prevColumn.maxWidth / 2;
    const leftBound = column.x - prevColumn.maxWidth / 2;
    const overlap = prevRightBound - leftBound;
    const oldSpacing = column.x - prevColumn.x;
    const newSpacing = oldSpacing + overlap;
    return Math.max(currentMax, (newSpacing / oldSpacing) * nodeSpacingMultiplier);
  }, 0);
}

// Canvas will be created on first use and cached for better performance
let textMeasureCanvas: HTMLCanvasElement;

/** Uses a canvas to measure the needed width to render a string using a given
 * font definition.
 */
export function measureText(fontSize: number, text: string) {
  if (!textMeasureCanvas) {
    textMeasureCanvas = document.createElement('canvas');
  }
  const context = textMeasureCanvas.getContext('2d');
  if (!context) {
    throw new Error('Unable to create canvas context for text measurement');
  }

  const fontString = `${fontSize}px sans-serif`;

  context.font = fontString;
  return context.measureText(text);
}

export interface AssignNodeTextWidthsResult<T> {
  // The resulting array of nodes. Nodes are modified in-place, so this is
  // provided for convenience.
  nodes: RenderableNode<T>[];
  // The maximum measured node width, useful for calculating the necessary
  // graph scale to prevent nodes from overlapping horizontally
  maxWidth: number;
}

/** For an array of nodes and a given font size, this will calculate the text
 * width for each node and assign it as a property `textWidth` to be used
 * when rendering the text rectangle for each node.
 */
export function assignNodeTextSizes<T>(
  input: d3.DierarchyPointNode<T>[],
  fontSize: number,
): AssignNodeTextWidthsResult<T> {
  const nodes = input as RenderableNode<T>[];
  const maxWidth = nodes.reduce((currentMax, node) => {
    const measured = measureText(fontSize, node.id);
    (node as RenderableNode<T>).textWidth = measured.width;
    return Math.max(currentMax, Math.ceil(measured.width));
  }, 0);
  return { maxWidth, nodes };
}

interface LayoutFunctions {
  coordFunction(): unknown;
  decrossFunction(): unknown;
  layeringFunction(): unknown;
}

function determineLayoutFunctions<T extends GraphInputNode>(nodes: T[]): LayoutFunctions {
  // This is the best balance between performance and aesthetics, so we don't
  // change it based on the graph.
  const layeringFunction = d3.layeringCoffmanGraham();
  // Using a simple criteria (node count) to switch off the expensive
  // algorithms. This isn't 100% guaranteed to be efficient (a graph with
  // an insane amount of edges of example), but should cover all practical cases
  const isLargeGraph = nodes.length > 10;
  const coordFunction = isLargeGraph ? d3.coordGreedy() : d3.coordVert();
  const decrossFunction = isLargeGraph ? d3.decrossTwoLayer() : d3.decrossOpt();
  return { coordFunction, decrossFunction, layeringFunction };
}

/** Assigns x/y positions to an array of connected graph nodes using a vertical
 * graph structure which optimizes for minimum link crossings. Will return the
 * root node of the arranged graph, which exposes functions for retrieving the
 * links and descendants.
 */
export function layoutGraph<T extends GraphInputNode>(
  input: T[],
  config: GraphConfig,
): GraphLayoutResult<T> {
  const {
    node: { fontSize, textPadding },
  } = config;
  const timer = createTimer();

  // The layout operations will modify the data in place, so make a copy
  const nodeList = cloneDeep(input);

  const dag = d3.dagStratify<T>()(nodeList);

  const { coordFunction, decrossFunction, layeringFunction } = determineLayoutFunctions(nodeList);

  const processedDag = d3
    .sugiyama<T>()
    .layering(layeringFunction)
    .size([layoutSize, layoutSize])
    .coord(coordFunction)
    .decross(decrossFunction)(dag);

  // We're rendering horizontally, so flip the x/y coordinates of nodes/links
  processedDag.descendants().forEach((node) => {
    const { x, y } = node;
    node.x = y;
    node.y = x;
  });

  processedDag.links().forEach((link) => {
    link.data.points.forEach((point) => {
      const { x, y } = point;
      point.x = y;
      point.y = x;
    });
  });

  // Measure the text size of each node and assign it a width to be used for
  // rendering
  const { nodes } = assignNodeTextSizes(processedDag.descendants(), fontSize);

  // Split the nodes into columns by x position and find the max width in each
  // column, to be used for calculating the ideal final size
  const columns = findColumnMeasurements(nodes);

  // Find the minimum scale needed to ensure nodes don't overlap
  const graphScale = findNeededHorizontalColumnScale(columns);

  const maxOuterColumnWidth = Math.max(
    columns[0].maxWidth / 2,
    columns[columns.length - 1].maxWidth / 2,
  );

  // we want enough padding on the sides to account for node
  // width and enough on top to account for node height, given that
  // the center of some nodes will be placed on the edges of the graph.
  // Each value is doubled to allow nodes room to grow or shrink up to 2x
  // without hitting the bounds
  const padding = {
    x: 2 * Math.ceil(maxOuterColumnWidth + config.node.strokeWidth),
    y: 2 * Math.ceil(fontSize + textPadding + config.node.strokeWidth),
  };

  // Now that we changed the graph size, we need to scale all the
  // node and link point positions to match
  processedDag.descendants().forEach((node) => {
    node.x = node.x * graphScale + padding.x;
    node.y = node.y * graphScale + padding.y;
  });

  processedDag.links().forEach((link) => {
    link.data.points.forEach((point) => {
      point.x = point.x * graphScale + padding.x;
      point.y = point.y * graphScale + padding.y;
    });
  });

  // Add the padding to the output graph dimensions
  const height = layoutSize * graphScale + padding.y * 2;
  const width = layoutSize * graphScale + padding.x * 2;

  log(`Layout time: ${timer.timeStringMS}`);
  return {
    height,
    width,
    rootNode: processedDag as RenderableNode<T>,
  };
}
