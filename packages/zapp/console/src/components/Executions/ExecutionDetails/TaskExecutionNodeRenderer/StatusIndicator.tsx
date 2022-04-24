import { makeStyles } from '@material-ui/core/styles';
import classnames from 'classnames';
import { NodeConfig, Point } from 'components/flytegraph/types';
import * as React from 'react';

const useStyles = makeStyles(() => ({
  pulse: {
    animation: '1200ms infinite alternate',
    animationName: 'pulse',
  },
  '@keyframes pulse': {
    '0%': {
      opacity: 0.35,
    },
    '80%': {
      opacity: 1,
    },
  },
}));

const statusSize = 11;

/** Renders an indicator for a node based on execution status */
export const StatusIndicator: React.FC<{
  color: string;
  config: NodeConfig;
  position: Point;
  pulse: boolean;
}> = ({ color, config, position, pulse }) => (
  <g transform={`translate(${position.x} ${position.y})`}>
    <circle
      className={classnames({ [useStyles().pulse]: pulse })}
      fill={color}
      stroke={config.strokeColor}
      strokeWidth={config.strokeWidth}
      r={statusSize / 2}
    />
  </g>
);
