import * as React from 'react';
import { useState } from 'react';
import { useMutation } from 'react-query';
import { makeStyles, Theme } from '@material-ui/core';
import classnames from 'classnames';
import { useSnackbar } from 'notistack';
import { Execution } from 'models/Execution/types';
import { ExecutionState } from 'models/Execution/enums';
import { updateExecution } from 'models/Execution/api';
import { ListRowProps } from 'react-virtualized';
import { isExecutionArchived } from '../../utils';
import { ExpandableExecutionError } from '../ExpandableExecutionError';
import { useExecutionTableStyles } from '../styles';
import {
    WorkflowExecutionColumnDefinition,
    WorkflowExecutionsTableState
} from '../types';
import { showOnHoverClass } from './cells';
import {
    useConfirmationSection,
    useWorkflowExecutionsTableColumns
} from './useWorkflowExecutionsTableColumns';
import t from './strings';

const useStyles = makeStyles((theme: Theme) => ({
    row: {
        paddingLeft: theme.spacing(2),
        // All children using the showOnHover class will be hidden until
        // the mouse enters the container
        [`& .${showOnHoverClass}`]: {
            opacity: 0
        },
        [`&:hover .${showOnHoverClass}`]: {
            opacity: 1
        }
    }
}));

export interface WorkflowExecutionRowProps extends Partial<ListRowProps> {
    showWorkflowName: boolean;
    errorExpanded?: boolean;
    execution: Execution;
    onExpandCollapseError?(expanded: boolean): void;
    state: WorkflowExecutionsTableState;
}

/** Renders a single `Execution` record as a row. Designed to be used as a child
 * of `WorkflowExecutionTable`.
 */
export const WorkflowExecutionRow: React.FC<WorkflowExecutionRowProps> = ({
    showWorkflowName,
    errorExpanded,
    execution,
    onExpandCollapseError,
    state,
    style
}) => {
    const { enqueueSnackbar } = useSnackbar();
    const tableStyles = useExecutionTableStyles();
    const styles = useStyles();

    const isArchived = isExecutionArchived(execution);
    const [hideItem, setHideItem] = useState<boolean>(false);
    const [isUpdating, setIsUpdating] = useState<boolean>(false);
    const [showConfirmation, setShowConfirmation] = useState<boolean>(false);

    const mutation = useMutation(
        (newState: ExecutionState) => updateExecution(execution.id, newState),
        {
            onMutate: () => setIsUpdating(true),
            onSuccess: () => {
                enqueueSnackbar(t('archiveSuccess', !isArchived), {
                    variant: 'success'
                });
                setHideItem(true);
                // ensure to collapse error info and re-calculate rows positions.
                onExpandCollapseError?.(false);
            },
            onError: () => {
                enqueueSnackbar(
                    `${mutation.error ?? t('archiveError', !isArchived)}`,
                    { variant: 'error' }
                );
            },
            onSettled: () => {
                setShowConfirmation(false);
                setIsUpdating(false);
            }
        }
    );

    const onArchiveConfirmClick = () => {
        mutation.mutate(
            isArchived
                ? ExecutionState.EXECUTION_ACTIVE
                : ExecutionState.EXECUTION_ARCHIVED
        );
    };

    const columns = useWorkflowExecutionsTableColumns({
        showWorkflowName,
        onArchiveClick: () => setShowConfirmation(true)
    });
    const confirmation = useConfirmationSection({
        isArchived,
        isLoading: isUpdating,
        onCancel: () => setShowConfirmation(false),
        onConfirmClick: onArchiveConfirmClick
    });
    const columnsWithApproval = [...columns.slice(0, -2), confirmation];

    // we show error info only on active items
    const { abortMetadata, error } = execution.closure;
    const showErrorInfo = !isArchived && (error || abortMetadata);

    const renderCell = ({
        className,
        key: columnKey,
        cellRenderer
    }: WorkflowExecutionColumnDefinition): JSX.Element => (
        <div
            key={columnKey}
            className={classnames(tableStyles.rowColumn, className)}
        >
            {cellRenderer({ execution, state })}
        </div>
    );

    if (hideItem) {
        return null;
    }

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
                {!showConfirmation
                    ? columns.map(renderCell)
                    : columnsWithApproval.map(renderCell)}
            </div>
            {showErrorInfo ? (
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
