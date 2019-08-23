import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { useCommonStyles } from 'components/common/styles';
import { Project } from 'models/Project';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes/routes';

const useStyles = makeStyles((theme: Theme) => ({
    projectCard: {
        textAlign: 'left',
        marginBottom: theme.spacing(2),
        width: theme.spacing(36)
    },
    projectTitle: {
        paddingBottom: 0
    }
}));

export interface ProjectListProps {
    error?: string;
    projects: Project[];
}

const DomainLinks: React.FC<{ project: Project }> = ({ project }) => {
    return (
        <section>
            {project.domains.map(({ id: domainId, name }, idx) => (
                <span key={domainId}>
                    {idx > 0 && <span>&nbsp;|&nbsp;</span>}
                    <Link
                        to={Routes.ProjectDetails.sections.workflows.makeUrl(
                            project.id,
                            domainId
                        )}
                    >
                        {name}
                    </Link>
                </span>
            ))}
        </section>
    );
};

const ProjectCard: React.FC<{ project: Project }> = ({ project }) => {
    const styles = useStyles();
    return (
        <Card className={styles.projectCard}>
            <CardHeader
                className={styles.projectTitle}
                title={project.name}
                titleTypographyProps={{ variant: 'body1' }}
            />
            <CardContent>
                <DomainLinks project={project} />
            </CardContent>
        </Card>
    );
};

/** Displays the available Projects and domains as a list of cards */
export const ProjectList: React.FC<ProjectListProps> = props => {
    const commonStyles = useCommonStyles();
    return (
        <ul className={commonStyles.listUnstyled}>
            {props.projects.map(project => (
                <li key={project.id}>
                    <ProjectCard project={project} />
                </li>
            ))}
        </ul>
    );
};
