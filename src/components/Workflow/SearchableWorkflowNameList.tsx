import { makeStyles } from '@material-ui/core/styles';
import DeviceHub from '@material-ui/icons/DeviceHub';
import classNames from 'classnames';
import { useNamedEntityListStyles } from 'components/common/SearchableNamedEntityList';
import { useCommonStyles } from 'components/common/styles';
import { separatorColor, primaryTextColor, workflowLabelColor } from 'components/Theme/constants';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes/routes';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import { Shimmer } from 'components/common/Shimmer';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { debounce } from 'lodash';
import { WorkflowListStructureItem } from './types';
import ProjectStatusBar from '../Project/ProjectStatusBar';
import { workflowNoInputsString } from '../Launch/LaunchForm/constants';
import { SearchableInput } from '../common/SearchableList';
import { useSearchableListState } from '../common/useSearchableListState';
import { useWorkflowInfoItem } from './useWorkflowInfoItem';

interface SearchableWorkflowNameItemProps {
  item: WorkflowListStructureItem;
}

interface SearchableWorkflowNameListProps {
  workflows: WorkflowListStructureItem[];
}

const useStyles = makeStyles(() => ({
  container: {
    padding: 13,
    paddingRight: 71,
  },
  itemContainer: {
    marginBottom: 15,
    borderRadius: 16,
    padding: '23px 30px',
    border: `1px solid ${separatorColor}`,
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
  },
  itemName: {
    display: 'flex',
    fontWeight: 600,
    color: primaryTextColor,
    marginBottom: 10,
  },
  itemDescriptionRow: {
    color: '#757575',
    marginBottom: 30,
    width: '100%',
  },
  itemIcon: {
    marginRight: 14,
    color: '#636379',
  },
  itemRow: {
    display: 'flex',
    marginBottom: 10,
    '&:last-child': {
      marginBottom: 0,
    },
    alignItems: 'center',
    width: '100%',
  },
  itemLabel: {
    width: 140,
    fontSize: 14,
    color: workflowLabelColor,
  },
  searchInputContainer: {
    padding: '0 13px',
    margin: '32px 0 23px',
  },
  w100: {
    flex: 1,
  },
}));

const padExecutions = (items: WorkflowExecutionPhase[]) => {
  if (items.length >= 10) {
    return items.slice(0, 10).reverse();
  }
  const emptyExecutions = new Array(10 - items.length).fill(WorkflowExecutionPhase.QUEUED);
  return [...items, ...emptyExecutions].reverse();
};

const padExecutionPaths = (items: WorkflowExecutionIdentifier[]) => {
  if (items.length >= 10) {
    return items
      .slice(0, 10)
      .map((id) => Routes.ExecutionDetails.makeUrl(id))
      .reverse();
  }
  const emptyExecutions = new Array(10 - items.length).fill(null);
  return [...items.map((id) => Routes.ExecutionDetails.makeUrl(id)), ...emptyExecutions].reverse();
};

/**
 * Renders individual searchable workflow item
 * @param item
 * @returns
 */
const SearchableWorkflowNameItem: React.FC<SearchableWorkflowNameItemProps> = React.memo(
  ({ item }) => {
    const commonStyles = useCommonStyles();
    const listStyles = useNamedEntityListStyles();
    const styles = useStyles();
    const { id, description } = item;
    const { data: workflow, isLoading } = useWorkflowInfoItem(id);

    return (
      <Link
        className={commonStyles.linkUnstyled}
        to={Routes.WorkflowDetails.makeUrl(id.project, id.domain, id.name)}
      >
        <div className={classNames(listStyles.searchResult, styles.itemContainer)}>
          <div className={styles.itemName}>
            <DeviceHub className={styles.itemIcon} />
            <div>{id.name}</div>
          </div>
          <div className={styles.itemDescriptionRow}>
            {description?.length ? description : 'This workflow has no description.'}
          </div>
          <div className={styles.itemRow}>
            <div className={styles.itemLabel}>Last execution time</div>
            <div className={styles.w100}>
              {isLoading ? (
                <Shimmer />
              ) : workflow.latestExecutionTime ? (
                workflow.latestExecutionTime
              ) : (
                <em>No executions found</em>
              )}
            </div>
          </div>
          <div className={styles.itemRow}>
            <div className={styles.itemLabel}>Last 10 executions</div>
            {isLoading ? (
              <Shimmer />
            ) : (
              <ProjectStatusBar
                items={padExecutions(workflow.executionStatus || [])}
                paths={padExecutionPaths(workflow.executionIds || [])}
              />
            )}
          </div>
          <div className={styles.itemRow}>
            <div className={styles.itemLabel}>Inputs</div>
            <div className={styles.w100}>
              {isLoading ? <Shimmer /> : workflow.inputs ?? <em>{workflowNoInputsString}</em>}
            </div>
          </div>
          <div className={styles.itemRow}>
            <div className={styles.itemLabel}>Outputs</div>
            <div className={styles.w100}>
              {isLoading ? <Shimmer /> : workflow?.outputs ?? <em>No output data found.</em>}
            </div>
          </div>
        </div>
      </Link>
    );
  },
);

/**
 * Renders a searchable list of Workflow names, with associated descriptions
 * @param workflows
 * @constructor
 */
export const SearchableWorkflowNameList: React.FC<SearchableWorkflowNameListProps> = ({
  workflows,
}) => {
  const styles = useStyles();
  const [search, setSearch] = React.useState('');
  const { results, setSearchString } = useSearchableListState({
    items: workflows,
    propertyGetter: ({ id }) => id.name,
  });

  const onSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const searchString = event.target.value;
    setSearch(searchString);
    const debouncedSearch = debounce(() => setSearchString(searchString), 1000);
    debouncedSearch();
  };
  const onClear = () => setSearch('');

  return (
    <>
      <SearchableInput
        onClear={onClear}
        onSearchChange={onSearchChange}
        variant="normal"
        value={search}
        className={styles.searchInputContainer}
        placeholder="Search Workflow Name"
      />
      <div className={styles.container}>
        {results.map(({ value }) => (
          <SearchableWorkflowNameItem
            item={value}
            key={`workflow-name-item-${value.id.domain}-${value.id.name}-${value.id.project}`}
          />
        ))}
      </div>
    </>
  );
};
