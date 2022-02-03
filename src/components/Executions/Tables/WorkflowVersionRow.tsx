import { makeStyles, TableCell, Theme } from '@material-ui/core';
import Radio from '@material-ui/core/Radio';
import classnames from 'classnames';
import * as React from 'react';
import { ListRowProps } from 'react-virtualized';
import { Workflow } from 'models/Workflow/types';
import { useWorkflowVersionsColumnStyles } from './styles';
import {
    WorkflowExecutionsTableState,
    WorkflowVersionColumnDefinition
} from './types';
import TableRow from '@material-ui/core/TableRow';

const useStyles = makeStyles((theme: Theme) => ({
    row: {
        cursor: 'pointer',
        height: theme.spacing(8)
    },
    cell: {
        padding: theme.spacing(1)
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
 * @param onClick
 * @param versionView
 * @param isChecked
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
    const versionTableStyles = useWorkflowVersionsColumnStyles();
    const styles = useStyles();

    return (
        <TableRow
            className={classnames(styles.row)}
            style={style}
            onClick={onClick}
        >
            {versionView && (
                <TableCell
                    classes={{
                        root: styles.cell
                    }}
                    className={versionTableStyles.columnRadioButton}
                >
                    <Radio checked={isChecked} />
                </TableCell>
            )}
            {columns.map(({ className, key: columnKey, cellRenderer }) => (
                <TableCell
                    key={columnKey}
                    classes={{
                        root: styles.cell
                    }}
                    className={classnames(className)}
                >
                    {cellRenderer({
                        workflow,
                        state
                    })}
                </TableCell>
            ))}
        </TableRow>
    );
};
