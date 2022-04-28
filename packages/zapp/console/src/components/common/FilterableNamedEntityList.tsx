import { Checkbox, debounce, FormControlLabel, FormGroup } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { NamedEntity } from 'models/Common/types';
import * as React from 'react';
import { NoResults } from './NoResults';
import { SearchableInput, SearchResult } from './SearchableList';
import { useCommonStyles } from './styles';
import { useSearchableListState } from './useSearchableListState';

export const useStyles = makeStyles((theme: Theme) => ({
  archiveCheckbox: {
    whiteSpace: 'nowrap',
  },
  container: {
    marginBottom: theme.spacing(2),
    width: '100%',
  },
  filterGroup: {
    display: 'flex',
    flexWrap: 'nowrap',
    flexDirection: 'row',
    margin: theme.spacing(4, 5, 2, 2),
  },
}));

type ItemRenderer = (item: SearchResult<NamedEntity>) => React.ReactNode;

interface SearchResultsProps {
  results: SearchResult<NamedEntity>[];
  renderItem: ItemRenderer;
}

export interface FilterableNamedEntityListProps {
  names: NamedEntity[];
  onArchiveFilterChange: (showArchievedItems: boolean) => void;
  showArchived: boolean;
  placeholder: string;
  archiveCheckboxLabel: string;
  renderItem: ItemRenderer;
}

const VARIANT = 'normal';

const SearchResults: React.FC<SearchResultsProps> = ({ renderItem, results }) => {
  const commonStyles = useCommonStyles();
  return results.length === 0 ? (
    <NoResults />
  ) : (
    <ul className={commonStyles.listUnstyled}>{results.map(renderItem)}</ul>
  );
};

/** Base component functionalityfor rendering NamedEntities (Workflow/Task/LaunchPlan) */
export const FilterableNamedEntityList: React.FC<FilterableNamedEntityListProps> = ({
  names,
  showArchived,
  renderItem,
  onArchiveFilterChange,
  placeholder,
  archiveCheckboxLabel,
}) => {
  const styles = useStyles();
  const [search, setSearch] = React.useState('');

  const { results, setSearchString } = useSearchableListState({
    items: names,
    propertyGetter: ({ id }) => id.name,
  });

  const onSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const searchString = event.target.value;
    setSearch(searchString);
    const debouncedSearch = debounce(() => setSearchString(searchString), 1000);
    debouncedSearch();
  };

  const onClear = () => setSearch('');

  const renderItems = (results: SearchResult<NamedEntity>[]) => (
    <SearchResults results={results} renderItem={renderItem} />
  );

  return (
    <div className={styles.container}>
      <FormGroup className={styles.filterGroup}>
        <SearchableInput
          onClear={onClear}
          onSearchChange={onSearchChange}
          placeholder={placeholder}
          value={search}
          variant={VARIANT}
        />
        <FormControlLabel
          className={styles.archiveCheckbox}
          control={
            <Checkbox
              checked={showArchived}
              onChange={(_, checked) => onArchiveFilterChange(checked)}
            />
          }
          label={archiveCheckboxLabel}
        />
      </FormGroup>
      {renderItems(results)}
    </div>
  );
};
