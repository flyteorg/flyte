import { Typography } from '@material-ui/core';
import classnames from 'classnames';
import { isFunction } from 'common/typeCheckers';
import * as React from 'react';
import { useExecutionTableStyles } from './styles';
import { ColumnDefinition } from './types';

/** Layout/rendering logic for the header row of an ExecutionsTable */
export const ExecutionsTableHeader: React.FC<{
  columns: ColumnDefinition<any>[];
  scrollbarPadding?: number;
  versionView?: boolean;
}> = ({ columns, scrollbarPadding = 0, versionView = false }) => {
  const tableStyles = useExecutionTableStyles();
  const scrollbarSpacer = scrollbarPadding > 0 ? <div style={{ width: scrollbarPadding }} /> : null;
  return (
    <div className={tableStyles.headerRow}>
      {versionView && (
        <div className={classnames(tableStyles.headerColumn, tableStyles.headerColumnVersion)}>
          &nbsp;
        </div>
      )}
      {columns.map(({ key, label, className }) => {
        const labelContent = isFunction(label) ? (
          React.createElement(label)
        ) : (
          <Typography variant="overline">{label}</Typography>
        );
        return (
          <div key={key} className={classnames(tableStyles.headerColumn, className)}>
            {labelContent}
          </div>
        );
      })}
      {scrollbarSpacer}
    </div>
  );
};
