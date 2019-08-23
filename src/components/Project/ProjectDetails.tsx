import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData, withRouteParams } from 'components/common';
import { useProject, useQueryState } from 'components/hooks';
import { Project } from 'models';
import * as React from 'react';
import { ProjectWorkflows } from './ProjectWorkflows';

const useStyles = makeStyles((theme: Theme) => ({
    tab: {
        textTransform: 'capitalize'
    },
    tabs: {
        borderBottom: `1px solid ${theme.palette.divider}`
    }
}));

export interface ProjectDetailsRouteParams {
    projectId: string;
}
export type ProjectDetailsProps = ProjectDetailsRouteParams;

const ProjectWorkflowsByDomain: React.FC<{ project: Project }> = ({
    project
}) => {
    const styles = useStyles();
    const { params, setQueryState } = useQueryState<{ domain: string }>();
    if (project.domains.length === 0) {
        throw new Error('No domains exist for this project');
    }
    const domainId = params.domain || project.domains[0].id;
    const handleTabChange = (event: React.ChangeEvent<{}>, tabId: string) =>
        setQueryState({
            domain: tabId
        });
    return (
        <>
            <Tabs
                className={styles.tabs}
                onChange={handleTabChange}
                value={domainId}
            >
                {project.domains.map(({ id, name }) => (
                    <Tab
                        className={styles.tab}
                        key={id}
                        value={id}
                        label={name}
                    />
                ))}
            </Tabs>
            <ProjectWorkflows projectId={project.id} domainId={domainId} />
        </>
    );
};

/** The view component for the Project landing page */
export const ProjectDetailsContainer: React.FC<ProjectDetailsRouteParams> = ({
    projectId
}) => {
    const project = useProject(projectId);
    return (
        <WaitForData {...project}>
            {() => {
                return <ProjectWorkflowsByDomain project={project.value} />;
            }}
        </WaitForData>
    );
};

export const ProjectDetails = withRouteParams<ProjectDetailsRouteParams>(
    ProjectDetailsContainer
);
