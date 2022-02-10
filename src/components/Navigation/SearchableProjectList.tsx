import { Fade, Tooltip, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { SearchableList, SearchResult } from 'components/common/SearchableList';
import { useCommonStyles } from 'components/common/styles';
import { defaultProjectDescription } from 'components/SelectProject/constants';
import { Project } from 'models/Project/types';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        marginBottom: theme.spacing(2),
        width: '100%'
    },
    itemName: {
        flex: '1 0 0',
        fontWeight: 'bold'
    },
    noResults: {
        color: theme.palette.text.disabled,
        display: 'flex',
        justifyContent: 'center',
        marginTop: theme.spacing(4)
    },
    searchResult: {
        alignItems: 'center',
        borderLeft: '4px solid transparent',
        cursor: 'pointer',
        display: 'flex',
        flexDirection: 'row',
        height: theme.spacing(5),
        padding: `0 ${theme.spacing(1)}px`,
        width: '100%',
        '&:hover': {
            borderColor: theme.palette.primary.main
        },
        '& mark': {
            backgroundColor: 'unset',
            color: theme.palette.primary.main,
            fontWeight: 'bold'
        }
    }
}));

type ProjectSelectedCallback = (project: Project) => void;

const NoResults: React.FC = () => (
    <Typography
        className={useStyles().noResults}
        variant="body2"
        component="div"
    >
        No matching results
    </Typography>
);

interface SearchResultsProps {
    onProjectSelected: ProjectSelectedCallback;
    results: SearchResult<Project>[];
}
const SearchResults: React.FC<SearchResultsProps> = ({
    onProjectSelected,
    results
}) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    return results.length === 0 ? (
        <NoResults />
    ) : (
        <ul className={commonStyles.listUnstyled}>
            {results.map(({ content, value }) => (
                <Tooltip
                    TransitionComponent={Fade}
                    key={value.id}
                    placement="bottom-end"
                    enterDelay={500}
                    title={
                        <Typography variant="body1">
                            <div className={commonStyles.textMonospace}>
                                {value.id}
                            </div>
                            <div>
                                <em>
                                    {value.description ||
                                        defaultProjectDescription}
                                </em>
                            </div>
                        </Typography>
                    }
                >
                    <div
                        className={styles.searchResult}
                        onClick={onProjectSelected.bind(null, value)}
                    >
                        <div
                            className={classnames(
                                styles.itemName,
                                commonStyles.textWrapped
                            )}
                        >
                            {content}
                        </div>
                    </div>
                </Tooltip>
            ))}
        </ul>
    );
};

export interface SearchableProjectListProps {
    onProjectSelected: ProjectSelectedCallback;
    projects: Project[];
}
/** Given a list of Projects, renders a searchable list of items which
 * navigate to the details page for the project on click
 */
export const SearchableProjectList: React.FC<SearchableProjectListProps> = ({
    onProjectSelected,
    projects
}) => {
    const styles = useStyles();

    const renderItems = (results: SearchResult<Project>[]) => (
        <SearchResults
            onProjectSelected={onProjectSelected}
            results={results}
        />
    );

    return (
        <div className={styles.container}>
            <SearchableList
                items={projects}
                placeholder="Filter Projects"
                propertyGetter="name"
                renderContent={renderItems}
                variant="minimal"
            />
        </div>
    );
};
