import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ChevronRight from '@material-ui/icons/ChevronRight';
import { SearchableList, SearchResult } from 'components/common/SearchableList';
import { useCommonStyles } from 'components/common/styles';
import { listhoverColor, separatorColor } from 'components/Theme';
import { NamedEntityIdentifier } from 'models';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        marginBottom: theme.spacing(2),
        width: '100%'
    },
    itemName: {
        flex: '1 0 auto'
    },
    itemChevron: {
        color: theme.palette.grey[500],
        flex: '0 0 auto'
    },
    noResults: {
        color: theme.palette.text.disabled,
        display: 'flex',
        justifyContent: 'center',
        marginTop: theme.spacing(6)
    },
    searchResult: {
        alignItems: 'center',
        borderBottom: `1px solid ${separatorColor}`,
        display: 'flex',
        flexDirection: 'row',
        height: theme.spacing(7),
        padding: `0 ${theme.spacing(3)}px`,
        '&:first-of-type': {
            borderTop: `1px solid ${separatorColor}`
        },
        '&:hover': {
            backgroundColor: listhoverColor
        },
        '& mark': {
            backgroundColor: 'unset',
            color: theme.palette.primary.main,
            fontWeight: 'bold'
        }
    }
}));

interface SearchableWorkflowId extends NamedEntityIdentifier {
    id: string;
}

const workflowIdKey = ({ domain, name, project }: NamedEntityIdentifier) =>
    `${domain}/${name}/${project}`;

const NoResults: React.FC = () => (
    <Typography className={useStyles().noResults} variant="h6" component="div">
        No matching results
    </Typography>
);

interface SearchResultsProps {
    results: SearchResult<SearchableWorkflowId>[];
}
const SearchResults: React.FC<SearchResultsProps> = ({ results }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    return results.length === 0 ? (
        <NoResults />
    ) : (
        <ul className={commonStyles.listUnstyled}>
            {results.map(({ id, content, value }) => (
                <Link
                    key={id}
                    className={commonStyles.linkUnstyled}
                    to={Routes.WorkflowDetails.makeUrl(
                        value.project,
                        value.domain,
                        value.name
                    )}
                >
                    <div className={styles.searchResult}>
                        <div className={styles.itemName}>{content}</div>
                        <ChevronRight className={styles.itemChevron} />
                    </div>
                </Link>
            ))}
        </ul>
    );
};

export interface SearchableWorkflowIdListProps {
    workflowIds: NamedEntityIdentifier[];
}
/** Given a list of WorkflowIds, renders a searchable list of items which
 * navigate to the WorkflowDetails page on click
 */
export const SearchableWorkflowIdList: React.FC<
    SearchableWorkflowIdListProps
> = ({ workflowIds }) => {
    const styles = useStyles();
    const searchValues = workflowIds.map(workflowId => ({
        ...workflowId,
        id: workflowIdKey(workflowId)
    }));

    const renderItems = (results: SearchResult<SearchableWorkflowId>[]) => (
        <SearchResults results={results} />
    );

    return (
        <div className={styles.container}>
            <SearchableList
                items={searchValues}
                propertyGetter="name"
                renderContent={renderItems}
            />
        </div>
    );
};
