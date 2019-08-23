import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { getScheduleFrequencyString } from 'common/formatters';
import { WaitForData } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { useWorkflowSchedules } from 'components/hooks';
import { LaunchPlan, NamedEntityIdentifier } from 'models';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
    schedulesContainer: {
        marginTop: theme.spacing(1)
    }
}));

const noSchedulesString = 'This workflow has no schedules.';

const RenderSchedules: React.FC<{ launchPlans: LaunchPlan[] }> = ({
    launchPlans
}) => {
    const commonStyles = useCommonStyles();
    if (launchPlans.length === 0) {
        return (
            <Typography variant="body2" className={commonStyles.hintText}>
                {noSchedulesString}
            </Typography>
        );
    }

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

export const WorkflowSchedules: React.FC<{
    workflowId: NamedEntityIdentifier;
}> = ({ workflowId }) => {
    const styles = useStyles();
    const scheduledLaunchPlans = useWorkflowSchedules(workflowId);
    return (
        <>
            <WaitForData {...scheduledLaunchPlans} spinnerVariant="none">
                <Typography variant="h6">Schedules</Typography>
                <div className={styles.schedulesContainer}>
                    <RenderSchedules launchPlans={scheduledLaunchPlans.value} />
                </div>
            </WaitForData>
        </>
    );
};
