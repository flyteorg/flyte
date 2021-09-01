import { makeStyles, Theme } from '@material-ui/core';
import classnames from 'classnames';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { ListRowProps } from 'react-virtualized';
import { ExpandableExecutionError } from './ExpandableExecutionError';
import { useExecutionTableStyles } from './styles';
import {
    WorkflowExecutionColumnDefinition,
    WorkflowExecutionsTableState
} from './types';

const useStyles = makeStyles((theme: Theme) => ({
    row: {
        paddingLeft: theme.spacing(2)
    }
}));

export interface WorkflowExecutionRowProps extends Partial<ListRowProps> {
    columns: WorkflowExecutionColumnDefinition[];
    errorExpanded?: boolean;
    execution: Execution;
    onExpandCollapseError?(expanded: boolean): void;
    state: WorkflowExecutionsTableState;
}

/** Renders a single `Execution` record as a row. Designed to be used as a child
 * of `WorkflowExecutionTable`.
 */
export const WorkflowExecutionRow: React.FC<WorkflowExecutionRowProps> = ({
    columns,
    errorExpanded,
    execution,
    onExpandCollapseError,
    state,
    style
}) => {
    const { abortMetadata, error } = execution.closure;
    const tableStyles = useExecutionTableStyles();
    const styles = useStyles();

    return (
        <div
            className={classnames(
                tableStyles.row,
                styles.row,
                tableStyles.borderBottom
            )}
            style={style}
        >
            <div className={tableStyles.rowColumns}>
                {columns.map(({ className, key: columnKey, cellRenderer }) => (
                    <div
                        key={columnKey}
                        className={classnames(tableStyles.rowColumn, className)}
                    >
                        {cellRenderer({
                            execution,
                            state
                        })}
                    </div>
                ))}
            </div>
            {error || abortMetadata ? (
                <ExpandableExecutionError
                    onExpandCollapse={onExpandCollapseError}
                    initialExpansionState={errorExpanded}
                    error={error}
                    abortMetadata={abortMetadata ?? undefined}
                />
            ) : null}
        </div>
    );
};
