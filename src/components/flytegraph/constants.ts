import { colors } from './theme';
import { GraphConfig } from './types';

export const componentIds = {
  arrowhead: 'arrowhead',
};

// This is the grid size used for layout with d3-dag. d3-dag will arrange all of
// the nodes in a square of this size, with x/y values in range [0, layoutSize]
export const layoutSize = 1000;

// Adjusts the amount of room between edges of adjacent nodes.
// Greater values lead to more space.
export const nodeSpacingMultiplier = 1.5;

export const defaultGraphConfig: GraphConfig = {
  node: {
    cornerRounding: 10,
    fillColor: colors.white,
    fontSize: 14,
    selectStrokeColor: colors.blue,
    selectStrokeWidth: 4,
    strokeColor: colors.darkGray1,
    strokeWidth: 2,
    textPadding: 10,
  },
  nodeLink: {
    strokeColor: colors.gray4,
    strokeWidth: 1,
  },
};

export const logPrefix = 'flytegraph';
