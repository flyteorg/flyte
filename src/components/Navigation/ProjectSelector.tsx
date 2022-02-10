import ButtonBase from '@material-ui/core/ButtonBase';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ExpandMore from '@material-ui/icons/ExpandMore';
import classnames from 'classnames';
import { KeyCodes } from 'common/constants';
import { useCommonStyles } from 'components/common/styles';
import { listhoverColor } from 'components/Theme/constants';
import { Project } from 'models/Project/types';
import * as React from 'react';
import { SearchableProjectList } from './SearchableProjectList';

const expanderGridHeight = 12;

const useStyles = makeStyles((theme: Theme) => ({
    expander: {
        alignItems: 'center',
        borderBottom: `${theme.spacing(1)}px solid ${listhoverColor}`,
        display: 'flex',
        flex: '0 0 auto',
        flexDirection: 'row',
        height: theme.spacing(expanderGridHeight),
        padding: theme.spacing(2),
        width: '100%',
        '&.expanded': {
            backgroundColor: listhoverColor
        }
    },
    expandIcon: {
        color: theme.palette.grey[500],
        flex: '0 0 auto'
    },
    header: {
        flex: '1 0 0',
        textAlign: 'left'
    },
    listContainer: {
        backgroundColor: theme.palette.background.default,
        bottom: 0,
        position: 'absolute',
        overflowY: 'scroll',
        top: theme.spacing(expanderGridHeight),
        width: '100%'
    }
}));

export interface ProjectSelectorProps {
    selectedProject: Project;
    projects: Project[];
    onProjectSelected: (project: Project) => void;
}

/** A complex selector that shows the current project when collapsed, and
 * renders a searchable list of projects when expanded.
 */
export const ProjectSelector: React.FC<ProjectSelectorProps> = ({
    projects,
    selectedProject,
    onProjectSelected
}) => {
    const styles = useStyles();
    const [expanded, setExpanded] = React.useState(false);
    const commonStyles = useCommonStyles();

    const onToggleExpanded = () => setExpanded(!expanded);
    const onSelect = (project: Project) => {
        setExpanded(false);
        onProjectSelected(project);
    };
    const onKeyDown = ({ keyCode }: React.KeyboardEvent) => {
        if (keyCode === KeyCodes.ESCAPE) {
            setExpanded(false);
        }
    };

    return (
        <div onKeyDownCapture={onKeyDown}>
            <ButtonBase
                disableRipple={true}
                disableTouchRipple={true}
                className={classnames(styles.expander, { expanded })}
                onClick={onToggleExpanded}
            >
                <header className={styles.header}>
                    <div className={commonStyles.microHeader}>PROJECT</div>
                    <div
                        className={classnames(
                            commonStyles.mutedHeader,
                            commonStyles.textWrapped
                        )}
                    >
                        {selectedProject.name}
                    </div>
                </header>
                <ExpandMore fontSize="large" className={styles.expandIcon} />
            </ButtonBase>
            {expanded && (
                <div className={styles.listContainer}>
                    <SearchableProjectList
                        onProjectSelected={onSelect}
                        projects={projects}
                    />
                </div>
            )}
        </div>
    );
};
