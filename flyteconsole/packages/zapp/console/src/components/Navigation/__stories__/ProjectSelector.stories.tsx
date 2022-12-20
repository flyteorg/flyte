import { makeStyles, Theme } from '@material-ui/core/styles';
import { storiesOf } from '@storybook/react';
import { sideNavGridWidth } from 'common/layout';
import { separatorColor } from 'components/Theme/constants';
import { Project } from 'models/Project/types';
import { createMockProjects } from 'models/__mocks__/projectData';
import * as React from 'react';
import { ProjectSelector } from '../ProjectSelector';

const stories = storiesOf('Navigation', module);

const projects = createMockProjects();

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    borderRight: `1px solid ${separatorColor}`,
    display: 'flex',
    flexDirection: 'column',
    bottom: 0,
    left: 0,
    position: 'fixed',
    top: 0,
    width: theme.spacing(sideNavGridWidth),
  },
}));

stories.addDecorator((story) => <div className={useStyles().root}>{story()}</div>);

stories.add('ProjectSelector', () => {
  const [selectedProject, setSelectedProject] = React.useState(projects[0]);
  const onProjectSelected = (project: Project) => setSelectedProject(project);
  return (
    <ProjectSelector
      projects={projects}
      selectedProject={selectedProject}
      onProjectSelected={onProjectSelected}
    />
  );
});
