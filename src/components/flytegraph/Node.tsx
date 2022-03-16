import * as React from 'react';
import { NodeText as DefaultNodeText } from './NodeText';
import { NodeRendererProps } from './types';

const hoverScaleMultiplier = 1.1;
const selectedScaleMultiplier = 1.2;
const transitionConfig = '0.1s ease-in-out';

const nodeStyles: Record<string, React.CSSProperties> = {
  base: {
    transition: `transform ${transitionConfig}`,
    userSelect: 'none',
  },
  hovered: {
    transform: `scale(${hoverScaleMultiplier})`,
  },
  selected: {
    transform: `scale(${selectedScaleMultiplier})`,
  },
};
const nodeContainerStyles: React.CSSProperties = {
  cursor: 'pointer',
};

const selectionRectStyles: Record<string, React.CSSProperties> = {
  base: {
    transition: `opacity ${transitionConfig}`,
    opacity: 0,
  },
  visible: {
    opacity: 1,
  },
};

function getNodeStyles({ selected, hovered }: NodeRendererProps<any>): React.CSSProperties {
  return {
    ...nodeStyles.base,
    ...(hovered && nodeStyles.hovered),
    ...(selected && nodeStyles.selected),
  };
}

function getSelectionRectStyles({ selected }: NodeRendererProps<any>): React.CSSProperties {
  return {
    ...selectionRectStyles.base,
    ...(selected && selectionRectStyles.visible),
  };
}

/** Default renderer for a Node, which will render a simple rounded rectangle */
export const Node: React.FC<NodeRendererProps<any>> = (props) => {
  const {
    children,
    config,
    node,
    selected = false,
    hovered = false,
    onClick,
    onEnter,
    onLeave,
    textRenderer: NodeText = DefaultNodeText,
  } = props;
  const {
    cornerRounding,
    fontSize,
    fillColor,
    selectStrokeColor,
    selectStrokeWidth,
    strokeColor,
    textPadding,
    strokeWidth,
  } = config;

  const height = fontSize + textPadding * 2;
  const width = node.textWidth + textPadding * 2;

  const baseRectProps: React.SVGProps<SVGRectElement> = {
    height,
    width,
    rx: cornerRounding,
    ry: cornerRounding,
    x: -width / 2,
    y: -height / 2,
  };
  return (
    <g onClick={onClick} onMouseEnter={onEnter} onMouseLeave={onLeave} style={nodeContainerStyles}>
      <g style={getNodeStyles(props)}>
        <rect {...baseRectProps} fill={fillColor} stroke={strokeColor} strokeWidth={strokeWidth} />
        <rect
          {...baseRectProps}
          fill="none"
          stroke={selectStrokeColor}
          strokeWidth={selectStrokeWidth}
          style={getSelectionRectStyles(props)}
        />
        <NodeText {...props} />
        {children}
      </g>
    </g>
  );
};
