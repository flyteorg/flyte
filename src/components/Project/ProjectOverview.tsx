import { Domain, Project } from 'models';
import * as React from 'react';
import { ProjectRecentExecutions } from './ProjectRecentExecutions';
import { ProjectSections } from './ProjectSections';

interface ProjectOverviewProps {
    domain: Domain;
    project: Project;
}

export const ProjectOveriew: React.FC<ProjectOverviewProps> = ({
    domain,
    project
}) => {
    return (
        <>
            <ProjectRecentExecutions project={project.id} domain={domain.id} />
            <ProjectSections project={project} domain={domain} />
        </>
    );
};
