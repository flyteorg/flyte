import * as React from 'react';

import { makeStyles, Theme } from '@material-ui/core/styles';
import { SectionHeader } from 'components/common';
import { flexWrap } from 'components/common/styles';
import { Domain, Project } from 'models';
import { ProjectRouteGenerator, Routes } from 'routes';
import { ProjectSectionCard } from './ProjectSectionCard';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        ...flexWrap(theme.spacing(2))
    }
}));

interface ProjectSection {
    description: string;
    label: string;
    // Will be combined with the current project/domain to create a full link
    makeUrl: ProjectRouteGenerator;
}
const projectSections: ProjectSection[] = [
    {
        label: 'Executions',
        description: 'View executions belonging to this project',
        makeUrl: Routes.ProjectExecutions.makeUrl
    },
    {
        label: 'Workflows',
        description: 'Browse this project by individual workflow',
        makeUrl: Routes.ProjectWorkflows.makeUrl
    },
    {
        label: 'Launch Plans',
        description: 'Explore pre-configured execution templates',
        makeUrl: Routes.ProjectLaunchPlans.makeUrl
    },
    {
        label: 'Schedules',
        description: 'See the automated executions for this project',
        makeUrl: Routes.ProjectSchedules.makeUrl
    }
];

export interface ProjectSectionsProps {
    domain: Domain;
    project: Project;
}

export const ProjectSections: React.FC<ProjectSectionsProps> = ({
    domain,
    project
}) => {
    const styles = useStyles();
    return (
        <section>
            <SectionHeader title="Explore" />
            <div className={styles.container}>
                {projectSections.map(({ makeUrl, ...cardProps }) => (
                    <ProjectSectionCard
                        key={cardProps.label}
                        link={makeUrl(project.id, domain.id)}
                        {...cardProps}
                    />
                ))}
            </div>
        </section>
    );
};
