import { Link, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ArrowBack from '@material-ui/icons/ArrowBack';
import * as classnames from 'classnames';
import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { useLocationState } from 'components/hooks/useLocationState';
import { NavBarContent } from 'components/Navigation/NavBarContent';
import { interactiveTextDisabledColor, smallFontSize } from 'components/Theme';
import { Execution } from 'models';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes';
import { ExecutionInputsOutputsModal } from '../ExecutionInputsOutputsModal';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { RelaunchExecutionButton } from '../RelaunchExecution';
import { TerminateExecutionButton } from '../TerminateExecution';
import { executionIsTerminal } from '../utils';

const useStyles = makeStyles((theme: Theme) => {
    const actionsMinWidth = theme.spacing(34);
    const badgeWidth = theme.spacing(11);
    const maxDetailsWidth = `calc(100% - ${actionsMinWidth + badgeWidth}px)`;
    return {
        actions: {
            alignItems: 'center',
            display: 'flex',
            justifyContent: 'flex-end',
            flex: '1 0 auto',
            height: '100%',
            marginLeft: theme.spacing(2)
        },
        backLink: {
            color: 'inherit',
            marginRight: theme.spacing(1)
        },
        container: {
            alignItems: 'center',
            display: 'flex',
            flex: '1 1 auto',
            maxWidth: '100%'
        },
        detailsContainer: {
            alignItems: 'center',
            display: 'flex',
            flex: '0 1 auto',
            justifyContent: 'space-between',
            maxWidth: maxDetailsWidth
        },
        detailItem: {
            flexShrink: 0,
            marginLeft: theme.spacing(2)
        },
        detailLabel: {
            fontSize: smallFontSize,
            lineHeight: 1.25
        },
        detailValue: {
            fontSize: '0.875rem',
            fontWeight: 'bold',
            lineHeight: '1.1875rem'
        },
        inputsOutputsLink: {
            color: interactiveTextDisabledColor
        },
        actionButton: {
            marginLeft: theme.spacing(2)
        },
        title: {
            flex: '0 1 auto',
            overflow: 'hidden'
        },
        version: {
            flex: '0 1 auto',
            overflow: 'hidden'
        }
    };
});

interface DetailItem {
    className?: string;
    label: React.ReactNode;
    value: React.ReactNode;
}

/** Renders information about a given Execution into the NavBar */
export const ExecutionDetailsAppBarContent: React.FC<{
    execution: Execution;
}> = ({ execution }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const [showInputsOutputs, setShowInputsOutputs] = React.useState(false);

    const { domain, name, project } = execution.id;
    const { duration, startedAt, phase, workflowId } = execution.closure;

    const {
        backLink = Routes.WorkflowDetails.makeUrl(
            workflowId.project,
            workflowId.domain,
            workflowId.name
        )
    } = useLocationState();

    let modalContent: JSX.Element | null = null;
    if (showInputsOutputs) {
        const onClose = () => setShowInputsOutputs(false);
        modalContent = (
            <ExecutionInputsOutputsModal
                execution={execution}
                onClose={onClose}
            />
        );
    }
    const onClickShowInputsOutputs = () => setShowInputsOutputs(true);
    const actionContent = executionIsTerminal(execution) ? (
        <RelaunchExecutionButton
            className={styles.actionButton}
            execution={execution}
        />
    ) : (
        <TerminateExecutionButton className={styles.actionButton} />
    );

    const details: DetailItem[] = [
        {
            className: styles.title,
            label: `${project}/${domain}/${workflowId.name}`,
            value: name
        },
        { label: 'Domain', value: domain },
        {
            className: styles.version,
            label: 'Version',
            value: workflowId.version
        },
        {
            label: 'Time',
            value: startedAt ? formatDateUTC(timestampToDate(startedAt)) : ''
        },
        {
            label: 'Duration',
            value: duration ? protobufDurationToHMS(duration) : ''
        }
    ];

    return (
        <>
            <NavBarContent>
                <div className={styles.container}>
                    <RouterLink className={styles.backLink} to={backLink}>
                        <ArrowBack />
                    </RouterLink>
                    <ExecutionStatusBadge phase={phase} type="workflow" />
                    <div className={styles.detailsContainer}>
                        {details.map(({ className, label, value }, idx) => (
                            <div
                                className={classnames(
                                    styles.detailItem,
                                    className
                                )}
                                key={idx}
                            >
                                <Typography
                                    className={classnames(
                                        styles.detailLabel,
                                        commonStyles.truncateText
                                    )}
                                    variant="body1"
                                >
                                    {label}
                                </Typography>
                                <div
                                    className={classnames(
                                        styles.detailValue,
                                        commonStyles.truncateText
                                    )}
                                >
                                    {value}
                                </div>
                            </div>
                        ))}
                    </div>
                    <div className={styles.actions}>
                        <Link
                            className={styles.inputsOutputsLink}
                            component="button"
                            onClick={onClickShowInputsOutputs}
                            variant="body1"
                        >
                            View Inputs &amp; Outputs
                        </Link>
                        {actionContent}
                    </div>
                </div>
            </NavBarContent>
            {modalContent}
        </>
    );
};
