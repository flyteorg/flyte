import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { getScheduleFrequencyString, getScheduleOffsetString } from 'common/formatters';
import { useCommonStyles } from 'components/common/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useWorkflowSchedules } from 'components/hooks/useWorkflowSchedules';
import { ResourceIdentifier } from 'models/Common/types';
import { LaunchPlan } from 'models/Launch/types';
import * as React from 'react';
import { entityStrings } from './constants';
import t from './strings';

const useStyles = makeStyles((theme: Theme) => ({
  schedulesContainer: {
    marginTop: theme.spacing(1),
  },
}));

const RenderSchedules: React.FC<{
  launchPlans: LaunchPlan[];
}> = ({ launchPlans }) => {
  const commonStyles = useCommonStyles();
  return (
    <ul className={commonStyles.listUnstyled}>
      {launchPlans.map((launchPlan, idx) => {
        const { schedule } = launchPlan.spec.entityMetadata;
        const frequencyString = getScheduleFrequencyString(schedule);
        const offsetString = getScheduleOffsetString(schedule);
        const scheduleString = offsetString
          ? `${frequencyString} (offset by ${offsetString})`
          : frequencyString;
        return <li key={idx}>{scheduleString}</li>;
      })}
    </ul>
  );
};

export const EntitySchedules: React.FC<{
  id: ResourceIdentifier;
}> = ({ id }) => {
  const styles = useStyles();
  const commonStyles = useCommonStyles();
  const scheduledLaunchPlans = useWorkflowSchedules(id);
  return (
    <>
      <WaitForData {...scheduledLaunchPlans} spinnerVariant="none">
        <Typography variant="h6">{t('schedulesHeader')}</Typography>
        <div className={styles.schedulesContainer}>
          {scheduledLaunchPlans.value.length > 0 ? (
            <RenderSchedules launchPlans={scheduledLaunchPlans.value} />
          ) : (
            <Typography variant="body2" className={commonStyles.hintText}>
              {t('noSchedules', entityStrings[id.resourceType])}
            </Typography>
          )}
        </div>
      </WaitForData>
    </>
  );
};
