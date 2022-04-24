export const componentIds = {
  arrowhead: 'arrowhead',
};

// This is the grid size used for layout with d3-dag. d3-dag will arrange all of
// the nodes in a square of this size, with x/y values in range [0, layoutSize]
export const layoutSize = 1000;

// Adjusts the amount of room between edges of adjacent nodes.
// Greater values lead to more space.
export const nodeSpacingMultiplier = 1.5;
