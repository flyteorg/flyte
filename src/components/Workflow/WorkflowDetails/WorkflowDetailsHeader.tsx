import { makeStyles, Theme } from '@material-ui/core/styles';
import ArrowBack from '@material-ui/icons/ArrowBack';
import * as classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { Project } from 'models';
import { getProjectDomain } from 'models/Project/utils';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes';

const useStyles = makeStyles((theme: Theme) => ({
    headerContainer: {
        alignItems: 'center',
        display: 'flex',
        height: theme.spacing(5),
        marginTop: theme.spacing(2),
        width: '100%'
    },
    headerText: {
        margin: `0 ${theme.spacing(1)}px`
    },
    headerTextContainer: {
        display: 'flex',
        flex: '1 0 auto'
    }
}));

export interface WorkflowDetailsRouteParams {
    projectId: string;
    domainId: string;
    workflowName: string;
}
export type WorkflowDetailsProps = WorkflowDetailsRouteParams;

interface WorkflowDetailsHeaderProps {
    domainId: string;
    project: Project;
    workflowName: string;
}

/** Renders the workflow name and actions shown on the workflow details page */
export const WorkflowDetailsHeader: React.FC<WorkflowDetailsHeaderProps> = ({
    domainId,
    project,
    workflowName
}) => {
    const styles = useStyles();
    const commonStyles = useCommonStyles();
    const domain = getProjectDomain(project, domainId);
    const headerText = `${domain.name} / ${workflowName}`;
    return (
        <div className={styles.headerContainer}>
            <div
                className={classnames(
                    commonStyles.mutedHeader,
                    styles.headerTextContainer
                )}
            >
                <Link
                    className={commonStyles.linkUnstyled}
                    to={Routes.ProjectDetails.sections.workflows.makeUrl(
                        project.id,
                        domainId
                    )}
                >
                    <ArrowBack color="inherit" />
                </Link>
                <span className={styles.headerText}>{headerText}</span>
            </div>
        </div>
    );
};
