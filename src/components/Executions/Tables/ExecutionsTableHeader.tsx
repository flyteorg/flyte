import * as classnames from 'classnames';
import * as React from 'react';
import { useExecutionTableStyles } from './styles';
import { ColumnDefinition } from './types';

/** Layout/rendering logic for the header row of an ExecutionsTable */
export const ExecutionsTableHeader: React.FC<{
    columns: ColumnDefinition<any>[];
    scrollbarPadding?: number;
}> = ({ columns, scrollbarPadding = 0 }) => {
    const tableStyles = useExecutionTableStyles();
    const scrollbarSpacer =
        scrollbarPadding > 0 ? (
            <div style={{ width: scrollbarPadding }} />
        ) : null;
    return (
        <div className={tableStyles.headerRow}>
            {columns.map(({ key, label, className }) => (
                <div
                    key={key}
                    className={classnames(tableStyles.headerColumn, className)}
                >
                    {label}
                </div>
            ))}
            {scrollbarSpacer}
        </div>
    );
};
