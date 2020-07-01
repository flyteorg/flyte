import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { unknownValueString } from 'common/constants';
import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { secondaryBackgroundColor } from 'components/Theme';
import { Execution } from 'models';
import * as React from 'react';
import { ExpandableExecutionError } from '../Tables/ExpandableExecutionError';
import { ExecutionMetadataLabels } from './constants';

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
    const { duration, error, startedAt, workflowId } = execution.closure;
    const { systemMetadata } = execution.spec.metadata;
    const cluster = systemMetadata?.executionCluster ?? unknownValueString;

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
        }
    ];
    if (startedAt) {
        details.push({
            label: ExecutionMetadataLabels.time,
            value: formatDateUTC(timestampToDate(startedAt))
        });
    }
    if (duration) {
        details.push({
            label: ExecutionMetadataLabels.duration,
            value: protobufDurationToHMS(duration)
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
            </div>

            {error ? <ExpandableExecutionError error={error} /> : null}
        </div>
    );
};
