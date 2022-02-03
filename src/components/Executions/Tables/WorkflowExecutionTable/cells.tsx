import * as React from 'react';
import {
    Typography,
    IconButton,
    Button,
    CircularProgress
} from '@material-ui/core';
import ArchiveOutlined from '@material-ui/icons/ArchiveOutlined';
import UnarchiveOutline from '@material-ui/icons/UnarchiveOutlined';
import LaunchPlanIcon from '@material-ui/icons/AssignmentOutlined';
import InputOutputIcon from '@material-ui/icons/Tv';
import {
    formatDateLocalTimezone,
    formatDateUTC,
    millisecondsToHMS
} from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { getWorkflowExecutionTimingMS, isExecutionArchived } from '../../utils';

import { WorkflowExecutionLink } from '../WorkflowExecutionLink';
import { ExecutionStatusBadge } from 'components/Executions/ExecutionStatusBadge';
import { Execution } from 'models/Execution/types';
import { ExecutionState, WorkflowExecutionPhase } from 'models/Execution/enums';
import { WorkflowExecutionsTableState } from '../types';
import classnames from 'classnames';
import { useStyles } from './styles';
import t from './strings';

export function getExecutionIdCell(
    execution: Execution,
    className: string,
    showWorkflowName?: boolean
): React.ReactNode {
    const { startedAt, workflowId } = execution.closure;
    const isArchived = isExecutionArchived(execution);

    return (
        <>
            <WorkflowExecutionLink
                id={execution.id}
                color={isArchived ? 'disabled' : 'primary'}
            />
            <Typography
                className={className}
                variant="subtitle1"
                color="textSecondary"
            >
                {showWorkflowName
                    ? workflowId.name
                    : t('lastRunStartedAt', startedAt)}
            </Typography>
        </>
    );
}

export function getStatusCell(execution: Execution): React.ReactNode {
    const isArchived = isExecutionArchived(execution);
    const phase = execution.closure.phase ?? WorkflowExecutionPhase.UNDEFINED;

    return (
        <ExecutionStatusBadge
            phase={phase}
            type="workflow"
            disabled={isArchived}
        />
    );
}

export function getStartTimeCell(execution: Execution): React.ReactNode {
    const { startedAt } = execution.closure;

    if (!startedAt) {
        return null;
    }

    const startedAtDate = timestampToDate(startedAt);
    const isArchived = isExecutionArchived(execution);

    return (
        <>
            <Typography
                variant="body1"
                color={isArchived ? 'textSecondary' : 'inherit'}
            >
                {formatDateUTC(startedAtDate)}
            </Typography>
            <Typography variant="subtitle1" color="textSecondary">
                {formatDateLocalTimezone(startedAtDate)}
            </Typography>
        </>
    );
}

export function getDurationCell(execution: Execution): React.ReactNode {
    const isArchived = isExecutionArchived(execution);
    const timing = getWorkflowExecutionTimingMS(execution);

    return (
        <Typography
            variant="body1"
            color={isArchived ? 'textSecondary' : 'inherit'}
        >
            {timing !== null ? millisecondsToHMS(timing.duration) : ''}
        </Typography>
    );
}

export const showOnHoverClass = 'showOnHover';
export function getActionsCell(
    execution: Execution,
    state: WorkflowExecutionsTableState,
    showLaunchPlan: boolean,
    wrapperClassName: string,
    iconClassName: string,
    onArchiveClick?: () => void //(newState: ExecutionState) => void,
): React.ReactNode {
    const isArchived = isExecutionArchived(execution);
    const onClick = () => state.setSelectedIOExecution(execution);

    const getArchiveIcon = (isArchived: boolean) =>
        isArchived ? <UnarchiveOutline /> : <ArchiveOutlined />;

    return (
        <div className={classnames(wrapperClassName, showOnHoverClass)}>
            <IconButton
                size="small"
                title={t('inputOutputTooltip')}
                onClick={onClick}
                className={iconClassName}
            >
                <InputOutputIcon />
            </IconButton>
            {showLaunchPlan && (
                <IconButton
                    size="small"
                    title={t('launchPlanTooltip')}
                    className={iconClassName}
                    onClick={() => {
                        /* Not implemented */
                    }}
                >
                    <LaunchPlanIcon />
                </IconButton>
            )}
            {!!onArchiveClick && (
                <IconButton
                    size="small"
                    title={t('archiveActionString', isArchived)}
                    onClick={onArchiveClick}
                >
                    {getArchiveIcon(isArchived)}
                </IconButton>
            )}
        </div>
    );
}

/**
 * ApprovalDooubleCell - represents approval request to Archive/Cancel operation on specific execution
 */
export interface ApprovalDoubleCellProps {
    isArchived: boolean;
    isLoading: boolean;
    onCancel: () => void;
    onConfirmClick: (newState: ExecutionState) => void;
}

export function ApprovalDoubleCell(props: ApprovalDoubleCellProps) {
    const { isArchived, isLoading, onCancel, onConfirmClick } = props;
    const styles = useStyles();

    if (isLoading) {
        return (
            <div className={styles.actionProgress}>
                <CircularProgress size={24} />
            </div>
        );
    }

    return (
        <>
            <Button
                size="medium"
                variant="contained"
                color="primary"
                className={styles.confirmationButton}
                disableElevation
                onClick={() =>
                    onConfirmClick(
                        isArchived
                            ? ExecutionState.EXECUTION_ACTIVE
                            : ExecutionState.EXECUTION_ARCHIVED
                    )
                }
            >
                {t('archiveAction', isArchived)}
            </Button>
            <Button
                size="medium"
                variant="contained"
                color="inherit"
                className={styles.confirmationButton}
                disableElevation
                onClick={onCancel}
            >
                {t('cancelAction')}
            </Button>
        </>
    );
}
