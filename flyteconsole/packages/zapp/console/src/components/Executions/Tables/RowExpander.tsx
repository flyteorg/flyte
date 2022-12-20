import { IconButton } from '@material-ui/core';
import ChevronRight from '@material-ui/icons/ChevronRight';
import ExpandMore from '@material-ui/icons/ExpandMore';
import * as React from 'react';
import t from './strings';

/** A simple expand/collapse arrow to be rendered next to row items. */
export const RowExpander: React.FC<{
  expanded: boolean;
  onClick: () => void;
}> = ({ expanded, onClick }) => (
  <IconButton
    disableRipple={true}
    disableTouchRipple={true}
    size="small"
    title={t('expanderTitle', expanded)}
    onClick={(e: React.MouseEvent<HTMLElement>) => {
      // prevent the parent row body onClick event trigger
      e.stopPropagation();
      onClick();
    }}
  >
    {expanded ? <ExpandMore /> : <ChevronRight />}
  </IconButton>
);
