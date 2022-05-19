import { ContentContainer, ContentContainerProps } from 'components/common/ContentContainer';
import { withSideNavigation } from 'components/Navigation/withSideNavigation';
import * as React from 'react';
import { Route, Switch } from 'react-router-dom';
import { components } from './components';
import { Routes } from './routes';

function withContentContainer<P>(
  WrappedComponent: React.ComponentType<P>,
  contentContainerProps?: ContentContainerProps,
) {
  return (props: P) => (
    <ContentContainer center={true} {...contentContainerProps}>
      <WrappedComponent {...props} />
    </ContentContainer>
  );
}

export const ApplicationRouter: React.FC = () => (
  <>
    <Switch>
      <Route
        path={Routes.ExecutionDetails.path}
        component={withContentContainer(components.executionDetails, {
          center: false,
          noMargin: true,
        })}
      />
      <Route
        path={Routes.TaskDetails.path}
        component={withSideNavigation(components.taskDetails)}
      />
      <Route
        exact
        path={Routes.WorkflowDetails.path}
        component={withSideNavigation(components.workflowDetails)}
      />
      <Route
        path={Routes.EntityVersionDetails.path}
        component={withSideNavigation(components.entityVersionDetails)}
      />
      <Route
        path={Routes.ProjectDetails.path}
        component={withSideNavigation(components.projectDetails, {
          noMargin: true,
        })}
      />
      <Route
        path={Routes.SelectProject.path}
        exact={true}
        component={withContentContainer(components.selectProject)}
      />
      <Route component={withContentContainer(components.notFound)} />
    </Switch>
  </>
);
