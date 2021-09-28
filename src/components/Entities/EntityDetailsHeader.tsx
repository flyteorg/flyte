import { Button } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ArrowBack from '@material-ui/icons/ArrowBack';
import * as classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { ResourceIdentifier } from 'models/Common/types';
import { Project } from 'models/Project/types';
import { getProjectDomain } from 'models/Project/utils';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { launchStrings } from './constants';
import { backUrlGenerator } from './generators';

const useStyles = makeStyles((theme: Theme) => ({
    actionsContainer: {},
    headerContainer: {
        alignItems: 'center',
        display: 'flex',
        height: theme.spacing(5),
        justifyContent: 'space-between',
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

interface EntityDetailsHeaderProps {
    project: Project;
    id: ResourceIdentifier;
    launchable?: boolean;
    versionView: boolean;
    onClickLaunch?(): void;
}

/**
 * Renders the entity name and any applicable actions.
 * @param id
 * @param onClickLaunch
 * @param launchable
 * @param versionView
 * @param project
 * @constructor
 */
export const EntityDetailsHeader: React.FC<EntityDetailsHeaderProps> = ({
    id,
    onClickLaunch,
    launchable = false,
    versionView,
    project
}) => {
    const styles = useStyles();
    const commonStyles = useCommonStyles();
    const domain = getProjectDomain(project, id.domain);
    const headerText = `${domain.name} / ${id.name}`;
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
                    to={
                        versionView
                            ? Routes.WorkflowDetails.makeUrl(
                                  id.project,
                                  id.domain,
                                  id.name
                              )
                            : backUrlGenerator[id.resourceType](id)
                    }
                >
                    <ArrowBack color="inherit" />
                </Link>
                <span className={styles.headerText}>{headerText}</span>
            </div>
            <div className={styles.actionsContainer}>
                {launchable ? (
                    <Button
                        color="primary"
                        id="launch-workflow"
                        onClick={onClickLaunch}
                        variant="contained"
                    >
                        {launchStrings[id.resourceType]}
                    </Button>
                ) : null}
            </div>
        </div>
    );
};
