import { makeStyles, Theme } from '@material-ui/core';
import classnames from 'classnames';
import * as React from 'react';
import { ListRowProps } from 'react-virtualized';
import { Workflow } from 'models/Workflow/types';
import { useExecutionTableStyles } from './styles';
import {
    WorkflowExecutionsTableState,
    WorkflowVersionColumnDefinition
} from './types';

const useStyles = makeStyles((theme: Theme) => ({
    row: {
        paddingLeft: theme.spacing(2)
    }
}));

export interface WorkflowVersionRowProps extends Partial<ListRowProps> {
    columns: WorkflowVersionColumnDefinition[];
    workflow: Workflow;
    state: WorkflowExecutionsTableState;
}

/**
 * Renders a single `Workflow` record as a row. Designed to be used as a child
 * of `WorkflowVersionsTable`.
 * @param columns
 * @param workflow
 * @param state
 * @param style
 * @constructor
 */
export const WorkflowVersionRow: React.FC<WorkflowVersionRowProps> = ({
    columns,
    workflow,
    state,
    style
}) => {
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
                            workflow,
                            state
                        })}
                    </div>
                ))}
            </div>
        </div>
    );
};
