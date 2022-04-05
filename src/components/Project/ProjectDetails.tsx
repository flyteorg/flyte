import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common/WaitForData';
import { withRouteParams } from 'components/common/withRouteParams';
import { useProject } from 'components/hooks/useProjects';
import { useQueryState } from 'components/hooks/useQueryState';
import { Project } from 'models/Project/types';
import * as React from 'react';
import { Redirect, Route, Switch } from 'react-router';
import { Routes } from 'routes/routes';
import { ProjectDashboard } from './ProjectDashboard';
import { ProjectTasks } from './ProjectTasks';
import { ProjectWorkflows } from './ProjectWorkflows';

const useStyles = makeStyles((theme: Theme) => ({
  tab: {
    textTransform: 'capitalize',
  },
  tabs: {
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
}));

export interface ProjectDetailsRouteParams {
  projectId: string;
}
export type ProjectDetailsProps = ProjectDetailsRouteParams;

const entityTypeToComponent = {
  executions: ProjectDashboard,
  tasks: ProjectTasks,
  workflows: ProjectWorkflows,
};

const ProjectEntitiesByDomain: React.FC<{
  project: Project;
  entityType: 'executions' | 'tasks' | 'workflows';
}> = ({ entityType, project }) => {
  const styles = useStyles();
  const { params, setQueryState } = useQueryState<{ domain: string }>();
  if (project.domains.length === 0) {
    throw new Error('No domains exist for this project');
  }
  const domainId = params.domain || project.domains[0].id;
  const handleTabChange = (_event: React.ChangeEvent<unknown>, tabId: string) =>
    setQueryState({
      domain: tabId,
    });
  const EntityComponent = entityTypeToComponent[entityType];
  return (
    <>
      <Tabs className={styles.tabs} onChange={handleTabChange} value={domainId}>
        {project.domains.map(({ id, name }) => (
          <Tab className={styles.tab} key={id} value={id} label={name} />
        ))}
      </Tabs>
      <EntityComponent projectId={project.id} domainId={domainId} />
    </>
  );
};

const ProjectDashboardByDomain: React.FC<{ project: Project }> = ({ project }) => (
  <ProjectEntitiesByDomain project={project} entityType="executions" />
);

const ProjectWorkflowsByDomain: React.FC<{ project: Project }> = ({ project }) => (
  <ProjectEntitiesByDomain project={project} entityType="workflows" />
);

const ProjectTasksByDomain: React.FC<{ project: Project }> = ({ project }) => (
  <ProjectEntitiesByDomain project={project} entityType="tasks" />
);

/** The view component for the Project landing page */
export const ProjectDetailsContainer: React.FC<ProjectDetailsRouteParams> = ({ projectId }) => {
  const project = useProject(projectId);
  return (
    <WaitForData {...project}>
      {() => {
        return (
          <Switch>
            <Route path={Routes.ProjectDetails.sections.dashboard.path}>
              <ProjectDashboardByDomain project={project.value} />
            </Route>
            <Route path={Routes.ProjectDetails.sections.workflows.path}>
              <ProjectWorkflowsByDomain project={project.value} />
            </Route>
            <Route path={Routes.ProjectDetails.sections.tasks.path}>
              <ProjectTasksByDomain project={project.value} />
            </Route>
            <Redirect to={Routes.ProjectDetails.sections.workflows.makeUrl(projectId)} />
          </Switch>
        );
      }}
    </WaitForData>
  );
};

export const ProjectDetails = withRouteParams<ProjectDetailsRouteParams>(ProjectDetailsContainer);
