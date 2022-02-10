import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ArrowBack from '@material-ui/icons/ArrowBack';
import classnames from 'classnames';
import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { NavBarContent } from 'components/Navigation/NavBarContent';
import { smallFontSize } from 'components/Theme/constants';
import { TaskExecution } from 'models/Execution/types';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';

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

/** Renders information about a given TaskExecution into the NavBar */
export const TaskExecutionDetailsAppBarContent: React.FC<{
    taskExecution: TaskExecution;
}> = ({ taskExecution }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();

    const {
        nodeExecutionId: { executionId },
        taskId: { project, domain, name, version },
        retryAttempt
    } = taskExecution.id;
    const { duration, startedAt, phase } = taskExecution.closure;
    const details: DetailItem[] = [
        {
            className: styles.title,
            label: `${executionId.project}/${executionId.name}`,
            value: `${project}/${domain}/${name} (${retryAttempt})`
        },
        { label: 'Domain', value: executionId.domain },
        {
            className: styles.version,
            label: 'Version',
            value: version
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
                    <RouterLink
                        className={styles.backLink}
                        to={Routes.ExecutionDetails.makeUrl(executionId)}
                    >
                        <ArrowBack />
                    </RouterLink>
                    <ExecutionStatusBadge phase={phase} type="task" />
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
                </div>
            </NavBarContent>
            {/* {modalContent} */}
        </>
    );
};
