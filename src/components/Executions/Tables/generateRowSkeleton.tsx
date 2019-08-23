import * as classnames from 'classnames';
import * as React from 'react';
import reactLoadingSkeleton from 'react-loading-skeleton';
import { useExecutionTableStyles } from './styles';
import { ColumnDefinition } from './types';

const Skeleton = reactLoadingSkeleton;

/** Creates a component for a row of skeleton placeholders based on
 * a list of ColumnDefinitions.
 */
export function generateRowSkeleton(
    columns: ColumnDefinition<any>[]
): React.FC {
    return () => {
        const tableStyles = useExecutionTableStyles();
        return (
            <div className={tableStyles.row}>
                <div className={tableStyles.rowColumns}>
                    <div className={tableStyles.expander}>
                        <Skeleton />
                    </div>
                    {columns.map(
                        ({ className, key: columnKey, cellRenderer }) => (
                            <div
                                key={columnKey}
                                className={classnames(
                                    tableStyles.rowColumn,
                                    className
                                )}
                            >
                                <Skeleton />
                            </div>
                        )
                    )}
                </div>
            </div>
        );
    };
}
