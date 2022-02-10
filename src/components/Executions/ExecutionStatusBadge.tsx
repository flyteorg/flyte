import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import {
    bodyFontFamily,
    smallFontSize,
    statusColors
} from 'components/Theme/constants';
import {
    NodeExecutionPhase,
    TaskExecutionPhase,
    WorkflowExecutionPhase
} from 'models/Execution/enums';
import * as React from 'react';
import {
    getNodeExecutionPhaseConstants,
    getTaskExecutionPhaseConstants,
    getWorkflowExecutionPhaseConstants
} from './utils';

const useStyles = makeStyles((theme: Theme) => ({
    default: {
        alignItems: 'center',
        backgroundColor: theme.palette.common.white,
        borderRadius: '4px',
        color: theme.palette.text.primary,
        display: 'flex',
        flex: '0 0 auto',
        height: theme.spacing(3),
        fontSize: smallFontSize,
        justifyContent: 'center',
        textTransform: 'uppercase',
        width: theme.spacing(11)
    },
    root: {
        fontFamily: bodyFontFamily,
        fontWeight: 'bold'
    },
    text: {
        backgroundColor: 'inherit',
        border: 'none',
        marginTop: theme.spacing(1),
        textTransform: 'lowercase'
    }
}));

interface BaseProps {
    variant?: 'default' | 'text';
    disabled?: boolean;
}

interface WorkflowExecutionStatusBadgeProps extends BaseProps {
    phase: WorkflowExecutionPhase;
    type: 'workflow';
}

interface NodeExecutionStatusBadgeProps extends BaseProps {
    phase: NodeExecutionPhase;
    type: 'node';
}

interface TaskExecutionStatusBadgeProps extends BaseProps {
    phase: TaskExecutionPhase;
    type: 'task';
}

type ExecutionStatusBadgeProps =
    | WorkflowExecutionStatusBadgeProps
    | NodeExecutionStatusBadgeProps
    | TaskExecutionStatusBadgeProps;

function getPhaseConstants(
    type: 'workflow' | 'node' | 'task',
    phase: WorkflowExecutionPhase | NodeExecutionPhase | TaskExecutionPhase
) {
    if (type === 'task') {
        return getTaskExecutionPhaseConstants(phase as TaskExecutionPhase);
    }
    if (type === 'node') {
        return getNodeExecutionPhaseConstants(phase as NodeExecutionPhase);
    }
    return getWorkflowExecutionPhaseConstants(phase as WorkflowExecutionPhase);
}

/** Given a `closure.phase` value for a Workflow/Task/NodeExecution, will render
 * a badge with the proper text and styling to indicate the status (succeeded/
 * failed etc.)
 */
export const ExecutionStatusBadge: React.FC<ExecutionStatusBadgeProps> = ({
    phase,
    type,
    variant = 'default',
    disabled = false
}) => {
    const styles = useStyles();
    const style: React.CSSProperties = {};
    const { badgeColor, text, textColor } = getPhaseConstants(type, phase);
    if (variant === 'text') {
        style.color = textColor;
    } else {
        style.backgroundColor = disabled ? statusColors.UNKNOWN : badgeColor;
    }

    return (
        <div className={classnames(styles.root, styles[variant])} style={style}>
            {text}
        </div>
    );
};
