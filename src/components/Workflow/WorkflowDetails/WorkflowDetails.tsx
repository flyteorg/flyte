import { Dialog } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData, withRouteParams } from 'components/common';
import { useProject } from 'components/hooks';
import { LaunchWorkflowForm } from 'components/Launch/LaunchWorkflowForm/LaunchWorkflowForm';
import * as React from 'react';
import { WorkflowDescription } from './WorkflowDescription';
import { WorkflowDetailsHeader } from './WorkflowDetailsHeader';
import { WorkflowExecutions } from './WorkflowExecutions';
import { WorkflowSchedules } from './WorkflowSchedules';

const useStyles = makeStyles((theme: Theme) => ({
    metadataContainer: {
        display: 'flex',
        marginBottom: theme.spacing(5),
        marginTop: theme.spacing(2),
        width: '100%'
    },
    descriptionContainer: {
        flex: '2 1 auto',
        marginRight: theme.spacing(2)
    },
    executionsContainer: {
        display: 'flex',
        flex: '1 1 auto',
        flexDirection: 'column',
        margin: `0 -${theme.spacing(contentMarginGridUnits)}px`
    },
    schedulesContainer: {
        flex: '1 2 auto',
        marginRight: theme.spacing(30)
    }
}));

export interface WorkflowDetailsRouteParams {
    projectId: string;
    domainId: string;
    workflowName: string;
}
export type WorkflowDetailsProps = WorkflowDetailsRouteParams;

/** The view component for the Workflow landing page */
export const WorkflowDetailsContainer: React.FC<WorkflowDetailsRouteParams> = ({
    projectId,
    domainId,
    workflowName
}) => {
    const project = useProject(projectId);
    const styles = useStyles();
    const [showLaunchForm, setShowLaunchForm] = React.useState(false);
    const onLaunch = () => setShowLaunchForm(true);
    const onCancelLaunch = () => setShowLaunchForm(false);

    const workflowId = {
        project: projectId,
        domain: domainId,
        name: workflowName
    };
    return (
        <>
            <WaitForData {...project}>
                <WorkflowDetailsHeader
                    project={project.value}
                    domainId={domainId}
                    workflowName={workflowName}
                    onClickLaunch={onLaunch}
                />
                <div className={styles.metadataContainer}>
                    <div className={styles.descriptionContainer}>
                        <WorkflowDescription workflowId={workflowId} />
                    </div>
                    <div className={styles.schedulesContainer}>
                        <WorkflowSchedules workflowId={workflowId} />
                    </div>
                </div>
                <div className={styles.executionsContainer}>
                    <WorkflowExecutions workflowId={workflowId} />
                </div>
                <Dialog
                    scroll="paper"
                    maxWidth="sm"
                    fullWidth={true}
                    open={showLaunchForm}
                >
                    <LaunchWorkflowForm
                        onClose={onCancelLaunch}
                        workflowId={workflowId}
                    />
                </Dialog>
            </WaitForData>
        </>
    );
};

export const WorkflowDetails = withRouteParams<WorkflowDetailsRouteParams>(
    WorkflowDetailsContainer
);
