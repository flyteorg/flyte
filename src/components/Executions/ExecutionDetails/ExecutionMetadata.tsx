import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { dashedValueString } from 'common/constants';
import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { secondaryBackgroundColor } from 'components/Theme/constants';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { ExpandableExecutionError } from '../Tables/ExpandableExecutionError';
import { ExecutionMetadataLabels } from './constants';
import { ExecutionMetadataExtra } from './ExecutionMetadataExtra';

const useStyles = makeStyles((theme: Theme) => {
    return {
        container: {
            background: secondaryBackgroundColor,
            display: 'flex',
            flexDirection: 'column',
            position: 'relative'
        },
        detailsContainer: {
            alignItems: 'center',
            display: 'flex',
            flex: '0 1 auto',
            paddingTop: theme.spacing(3),
            paddingBottom: theme.spacing(2)
        },
        detailItem: {
            flexShrink: 0,
            marginLeft: theme.spacing(4)
        },
        expandCollapseButton: {
            transition: theme.transitions.create('transform'),
            '&.expanded': {
                transform: 'rotate(180deg)'
            }
        },
        expandCollapseContainer: {
            bottom: 0,
            position: 'absolute',
            right: theme.spacing(2),
            transform: 'translateY(100%)',
            zIndex: 1
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

/** Renders metadata details about a given Execution */
export const ExecutionMetadata: React.FC<{
    execution: Execution;
}> = ({ execution }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();

    const { domain } = execution.id;
    const {
        abortMetadata,
        duration,
        error,
        startedAt,
        workflowId
    } = execution.closure;
    const { referenceExecution, systemMetadata } = execution.spec.metadata;
    const cluster = systemMetadata?.executionCluster ?? dashedValueString;

    const details: DetailItem[] = [
        { label: ExecutionMetadataLabels.domain, value: domain },
        {
            className: styles.version,
            label: ExecutionMetadataLabels.version,
            value: workflowId.version
        },
        {
            label: ExecutionMetadataLabels.cluster,
            value: cluster
        },
        {
            label: ExecutionMetadataLabels.time,
            value: startedAt
                ? formatDateUTC(timestampToDate(startedAt))
                : dashedValueString
        },
        {
            label: ExecutionMetadataLabels.duration,
            value: duration
                ? protobufDurationToHMS(duration)
                : dashedValueString
        }
    ];

    if (referenceExecution != null) {
        details.push({
            label: ExecutionMetadataLabels.relatedTo,
            value: (
                <RouterLink
                    className={commonStyles.primaryLink}
                    to={Routes.ExecutionDetails.makeUrl(referenceExecution)}
                >
                    {referenceExecution.name}
                </RouterLink>
            )
        });
    }

    return (
        <div className={styles.container}>
            <div className={styles.detailsContainer}>
                {details.map(({ className, label, value }, idx) => (
                    <div
                        className={classnames(styles.detailItem, className)}
                        key={idx}
                    >
                        <Typography
                            className={commonStyles.truncateText}
                            variant="subtitle1"
                        >
                            {label}
                        </Typography>
                        <Typography
                            className={commonStyles.truncateText}
                            variant="h6"
                            data-testid={`metadata-${label}`}
                        >
                            {value}
                        </Typography>
                    </div>
                ))}
                <ExecutionMetadataExtra execution={execution} />
            </div>

            {error || abortMetadata ? (
                <ExpandableExecutionError
                    abortMetadata={abortMetadata ?? undefined}
                    error={error}
                />
            ) : null}
        </div>
    );
};
