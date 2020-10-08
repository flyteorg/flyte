import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { getScheduleFrequencyString } from 'common/formatters';
import { WaitForData } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { useWorkflowSchedules } from 'components/hooks';
import { LaunchPlan, ResourceIdentifier } from 'models';
import * as React from 'react';
import { noSchedulesStrings, schedulesHeader } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
    schedulesContainer: {
        marginTop: theme.spacing(1)
    }
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
                return <li key={idx}>{frequencyString}</li>;
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
                <Typography variant="h6">{schedulesHeader}</Typography>
                <div className={styles.schedulesContainer}>
                    {scheduledLaunchPlans.value.length > 0 ? (
                        <RenderSchedules
                            launchPlans={scheduledLaunchPlans.value}
                        />
                    ) : (
                        <Typography
                            variant="body2"
                            className={commonStyles.hintText}
                        >
                            {noSchedulesStrings[id.resourceType]}
                        </Typography>
                    )}
                </div>
            </WaitForData>
        </>
    );
};
