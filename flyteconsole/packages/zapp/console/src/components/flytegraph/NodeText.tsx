import * as contrast from 'contrast';
import * as React from 'react';
import { textColors } from './theme';
import { NodeTextRendererProps } from './types';

/** Renders the text content for an individual node. */
export const NodeText: React.FC<NodeTextRendererProps<any>> = ({ config, node }) => {
  const textColor = config.textColor ? config.textColor : textColors[contrast(config.fillColor)];
  return (
    <text
      fill={textColor}
      fontSize={config.fontSize}
      dominantBaseline="middle"
      lengthAdjust="spacingAndGlyphs"
      textAnchor="middle"
      textLength={node.textWidth}
      key={`nodeTitle-${node.id}`}
    >
      {node.id}
    </text>
  );
};
