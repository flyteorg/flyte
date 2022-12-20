import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { getLaunchPlan } from 'models/Launch/api';
import { LaunchPlanSpec } from 'models/Launch/types';
import { dashedValueString } from 'common/constants';
import { ExecutionMetadataLabels } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
  detailItem: {
    flexShrink: 0,
    marginLeft: theme.spacing(4),
  },
}));

interface DetailItem {
  className?: string;
  label: ExecutionMetadataLabels;
  value: React.ReactNode;
}

/**
 * Renders extra metadata details about a given Execution
 * @param execution
 * @constructor
 */
export const ExecutionMetadataExtra: React.FC<{
  execution: Execution;
}> = ({ execution }) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const {
    launchPlan: launchPlanId,
    maxParallelism,
    rawOutputDataConfig,
    securityContext,
    interruptible,
  } = execution.spec;

  const [launchPlanSpec, setLaunchPlanSpec] = React.useState<Partial<LaunchPlanSpec>>({});

  React.useEffect(() => {
    getLaunchPlan(launchPlanId).then(({ spec }) => setLaunchPlanSpec(spec));
  }, [launchPlanId]);

  const details: DetailItem[] = [
    {
      label: ExecutionMetadataLabels.iam,
      value: securityContext?.runAs?.iamRole || ExecutionMetadataLabels.securityContextDefault,
    },
    {
      label: ExecutionMetadataLabels.serviceAccount,
      value:
        securityContext?.runAs?.k8sServiceAccount || ExecutionMetadataLabels.securityContextDefault,
    },
    {
      label: ExecutionMetadataLabels.rawOutputPrefix,
      value:
        rawOutputDataConfig?.outputLocationPrefix ||
        launchPlanSpec?.rawOutputDataConfig?.outputLocationPrefix,
    },
    {
      label: ExecutionMetadataLabels.parallelism,
      value: maxParallelism,
    },
    {
      label: ExecutionMetadataLabels.interruptible,
      value: interruptible ? (interruptible.value ? 'true' : 'false') : dashedValueString,
    },
  ];

  return (
    <>
      {details.map(({ className, label, value }) => (
        <div className={classnames(styles.detailItem, className)} key={label}>
          <Typography className={commonStyles.truncateText} variant="subtitle1">
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
    </>
  );
};
