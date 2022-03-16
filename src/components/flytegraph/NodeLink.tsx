import * as d3Shape from 'd3-shape';
import * as React from 'react';

import { componentIds } from './constants';
import { NodeLinkRendererProps, Point } from './types';
import { getMidpoint } from './utils';

// Given a set of points, will generate path data for a curve through them
const generateLine: d3Shape.Line<Point> = d3Shape
  .line<Point>()
  .curve(d3Shape.curveCatmullRom)
  .x((d) => d.x)
  .y((d) => d.y);

/** The default NodeLink renderer. This component draws links
 * as Catmull-Rom curves in a <path> element using at least the start point
 * (source) and end point (target). Additional mid points are passed as
 * `link.data.points`. If no midpoints exist (such as for a straight line), this
 * component will calculate and insert a point to ensure at least one
 * directional marker is drawn along the line. */
export const NodeLink: React.FC<NodeLinkRendererProps<any>> = ({ config, link }) => {
  const { strokeColor, strokeWidth } = config;
  const {
    source,
    target,
    data: { points },
  } = link;

  // We need at least one midpoint for an arrow to be drawn.
  // For simple straight paths, no midpoints will have been
  // generated, so we must add one.
  const midpoints = points && points.length ? points : [getMidpoint(source, target)];
  const finalPoints: Point[] = [
    { x: source.x, y: source.y },
    ...midpoints,
    { x: target.x, y: target.y },
  ];

  return (
    <path
      stroke={strokeColor}
      strokeWidth={strokeWidth}
      key={`link-${source.id}-${target.id}`}
      markerMid={`url(#${componentIds.arrowhead})`}
      d={generateLine(finalPoints)!}
    />
  );
};
