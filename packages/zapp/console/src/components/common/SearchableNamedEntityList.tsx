import { makeStyles, Theme } from '@material-ui/core/styles';
import { listhoverColor, separatorColor } from 'components/Theme/constants';
import { NamedEntity } from 'models/Common/types';
import * as React from 'react';
import { NoResults } from './NoResults';
import { SearchableList, SearchResult } from './SearchableList';
import { useCommonStyles } from './styles';

export const useNamedEntityListStyles = makeStyles((theme: Theme) => ({
  container: {
    marginBottom: theme.spacing(2),
    width: '100%',
  },
  itemName: {
    flex: '1 1 auto',
    padding: `${theme.spacing(2)}px 0`,
  },
  itemChevron: {
    color: theme.palette.grey[500],
    flex: '0 0 auto',
  },
  searchResult: {
    alignItems: 'center',
    borderBottom: `1px solid ${separatorColor}`,
    display: 'flex',
    position: 'relative',
    flexDirection: 'row',
    padding: `0 ${theme.spacing(3)}px`,
    '&:first-of-type': {
      borderTop: `1px solid ${separatorColor}`,
    },
    '&:hover': {
      backgroundColor: listhoverColor,
    },
    '& mark': {
      backgroundColor: 'unset',
      color: theme.palette.primary.main,
      fontWeight: 'bold',
    },
  },
}));

export interface SearchableNamedEntity extends NamedEntity {
  key: string;
}

const nameKey = ({ id: { domain, name, project } }: NamedEntity) => `${domain}/${name}/${project}`;

type ItemRenderer = (item: SearchResult<SearchableNamedEntity>) => React.ReactNode;

interface SearchResultsProps {
  results: SearchResult<SearchableNamedEntity>[];
  renderItem: ItemRenderer;
}
const SearchResults: React.FC<SearchResultsProps> = ({ renderItem, results }) => {
  const commonStyles = useCommonStyles();
  return results.length === 0 ? (
    <NoResults />
  ) : (
    <ul className={commonStyles.listUnstyled}>{results.map(renderItem)}</ul>
  );
};

export interface SearchableNamedEntityListProps {
  names: NamedEntity[];
  renderItem: ItemRenderer;
}

const nameSearchPropertyGetter = ({ id }: SearchableNamedEntity) => id.name;
/** Base component functionalityfor rendering NamedEntities (Workflow/Task/LaunchPlan) */
export const SearchableNamedEntityList: React.FC<SearchableNamedEntityListProps> = ({
  names,
  renderItem,
}) => {
  const styles = useNamedEntityListStyles();
  const searchValues = names.map((name) => ({
    ...name,
    key: nameKey(name),
  }));

  const renderItems = (results: SearchResult<SearchableNamedEntity>[]) => (
    <SearchResults results={results} renderItem={renderItem} />
  );

  return (
    <div className={styles.container}>
      <SearchableList
        items={searchValues}
        propertyGetter={nameSearchPropertyGetter}
        renderContent={renderItems}
      />
    </div>
  );
};
