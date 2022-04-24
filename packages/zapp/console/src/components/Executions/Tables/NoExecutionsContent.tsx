import { Typography } from '@material-ui/core';
import { noExecutionsFoundString } from 'common/constants';
import * as React from 'react';
import { useExecutionTableStyles } from './styles';

type SizeVariant = 'small' | 'large';

/** A message to show as a placeholder when we have an empty list of executions */
export const NoExecutionsContent: React.FC<{ size?: SizeVariant }> = ({ size = 'small' }) => (
  <div className={useExecutionTableStyles().noRowsContent}>
    <Typography variant={size === 'large' ? 'h6' : 'body1'}>{noExecutionsFoundString}</Typography>
  </div>
);
