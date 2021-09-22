import { Dialog } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData } from 'components/common/WaitForData';
import { EntityDescription } from 'components/Entities/EntityDescription';
import { useProject } from 'components/hooks/useProjects';
import { LaunchForm } from 'components/Launch/LaunchForm/LaunchForm';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import * as React from 'react';
import { entitySections } from './constants';
import { EntityDetailsHeader } from './EntityDetailsHeader';
import { EntityExecutions } from './EntityExecutions';
import { EntitySchedules } from './EntitySchedules';
import { EntityVersions } from './EntityVersions';
import classNames from 'classnames';
import { StaticGraphContainer } from 'components/Workflow/StaticGraphContainer';
import { WorkflowId } from 'models/Workflow/types';
import { EntityExecutionsBarChart } from './EntityExecutionsBarChart';

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
    versionsContainer: {
        display: 'flex',
        flexDirection: 'column',
        height: theme.spacing(53)
    },
    versionView: {
        flex: '1 1 auto'
    },
    schedulesContainer: {
        flex: '1 2 auto',
        marginRight: theme.spacing(30)
    }
}));

export interface EntityDetailsProps {
    id: ResourceIdentifier;
    versionView?: boolean;
    showStaticGraph?: boolean;
}

function getLaunchProps(id: ResourceIdentifier) {
    if (id.resourceType === ResourceType.TASK) {
        return { taskId: id };
    }

    return { workflowId: id };
}

/**
 * A view which optionally renders description, schedules, executions, and a
 * launch button/form for a given entity. Note: not all components are suitable
 * for use with all entities (not all entities have schedules, for example).
 * @param id
 * @param versionView
 */
export const EntityDetails: React.FC<EntityDetailsProps> = ({
    id,
    versionView = false,
    showStaticGraph = false
}) => {
    const sections = entitySections[id.resourceType];
    const workflowId = id as WorkflowId;
    const project = useProject(id.project);
    const styles = useStyles();
    const [showLaunchForm, setShowLaunchForm] = React.useState(false);
    const onLaunch = () => setShowLaunchForm(true);
    const onCancelLaunch = () => setShowLaunchForm(false);

    return (
        <WaitForData {...project}>
            <EntityDetailsHeader
                project={project.value}
                id={id}
                launchable={!!sections.launch}
                onClickLaunch={onLaunch}
            />
            {!versionView && (
                <div className={styles.metadataContainer}>
                    {sections.description ? (
                        <div className={styles.descriptionContainer}>
                            <EntityDescription id={id} />
                        </div>
                    ) : null}
                    {sections.schedules ? (
                        <div className={styles.schedulesContainer}>
                            <EntitySchedules id={id} />
                        </div>
                    ) : null}
                </div>
            )}
            {sections.versions ? (
                <>
                    {showStaticGraph ? (
                        <StaticGraphContainer workflowId={workflowId} />
                    ) : null}
                    <div
                        className={classNames(styles.versionsContainer, {
                            [styles.versionView]: versionView
                        })}
                    >
                        <EntityVersions id={id} versionView={versionView} />
                    </div>
                </>
            ) : null}
            {!versionView && <EntityExecutionsBarChart id={id} />}
            {sections.executions && !versionView ? (
                <div className={styles.executionsContainer}>
                    <EntityExecutions id={id} />
                </div>
            ) : null}
            {sections.launch ? (
                <Dialog
                    scroll="paper"
                    maxWidth="sm"
                    fullWidth={true}
                    open={showLaunchForm}
                >
                    <LaunchForm
                        onClose={onCancelLaunch}
                        {...getLaunchProps(id)}
                    />
                </Dialog>
            ) : null}
        </WaitForData>
    );
};
