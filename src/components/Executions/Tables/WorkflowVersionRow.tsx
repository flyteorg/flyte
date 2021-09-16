import { makeStyles, Theme } from '@material-ui/core';
import Radio from '@material-ui/core/Radio';
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
        paddingLeft: theme.spacing(2),
        cursor: 'pointer'
    }
}));

export interface WorkflowVersionRowProps extends Partial<ListRowProps> {
    columns: WorkflowVersionColumnDefinition[];
    workflow: Workflow;
    state: WorkflowExecutionsTableState;
    onClick: (() => void) | undefined;
    versionView?: boolean;
    isChecked?: boolean;
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
    style,
    onClick,
    versionView = false,
    isChecked = false
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
            onClick={onClick}
        >
            <div className={tableStyles.rowColumns}>
                {versionView && <Radio checked={isChecked} />}
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
