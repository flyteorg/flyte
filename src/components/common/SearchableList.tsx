import { IconButton } from '@material-ui/core';
import InputAdornment from '@material-ui/core/InputAdornment';
import { makeStyles, Theme } from '@material-ui/core/styles';
import TextField, { TextFieldProps } from '@material-ui/core/TextField';
import Close from '@material-ui/icons/Close';
import Search from '@material-ui/icons/Search';
import { bodyFontSize } from 'components/Theme/constants';
import * as React from 'react';
import classNames from 'classnames';
import { PropertyGetter, SearchResult, useSearchableListState } from './useSearchableListState';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    padding: `0 ${theme.spacing(3)}px`,
    marginBottom: theme.spacing(0.5),
    width: '100%',
  },
  containerMinimal: {
    margin: theme.spacing(1),
  },
  minimalNotchedOutline: {
    borderRadius: 0,
  },
  minimalInput: {
    fontSize: bodyFontSize,
    padding: theme.spacing(1),
  },
}));

type SearchableListVariant = 'normal' | 'minimal';

export interface SearchableListProps<T> {
  /** Note that all items must have an id property! */
  items: T[];
  /** Text to show in the search box when no query has been entered */
  placeholder?: string;
  /** The name of the property on each item which is being used for search */
  propertyGetter: keyof T | PropertyGetter<T>;
  variant?: SearchableListVariant;
  renderContent(results: SearchResult<T>[]): JSX.Element;
}

/**
 * Show searchable text input field.
 * @param onClear
 * @param onSearchChange
 * @param placeholder
 * @param value
 * @param variant
 * @constructor
 */
export const SearchableInput: React.FC<{
  onClear: () => void;
  onSearchChange: React.ChangeEventHandler<HTMLInputElement>;
  placeholder?: string;
  variant: SearchableListVariant;
  value?: string;
  className?: string;
}> = ({ onClear, onSearchChange, placeholder, value, variant, className }) => {
  const styles = useStyles();
  const startAdornment = (
    <InputAdornment position="start">
      <Search />
    </InputAdornment>
  );

  const endAdornment = (
    <InputAdornment position="end">
      <IconButton aria-label="Clear search query" onClick={onClear} size="small">
        <Close />
      </IconButton>
    </InputAdornment>
  );

  const baseProps: TextFieldProps = {
    placeholder,
    value,
    autoFocus: true,
    dir: 'auto',
    fullWidth: true,
    inputProps: { role: 'search' },
    onChange: onSearchChange,
    variant: 'outlined',
  };
  switch (variant) {
    case 'normal':
      return (
        <div className={className ? classNames(styles.container, className) : styles.container}>
          <TextField {...baseProps} margin="dense" InputProps={{ endAdornment, startAdornment }} />
        </div>
      );
    case 'minimal':
      return (
        <div className={styles.containerMinimal}>
          <TextField
            {...baseProps}
            margin="none"
            InputProps={{
              classes: {
                input: styles.minimalInput,
                notchedOutline: styles.minimalNotchedOutline,
              },
            }}
          />
        </div>
      );
  }
};

/** Handles fuzzy search logic and filtering for a searchable list of items */
export const SearchableList = <T extends {}>(props: SearchableListProps<T>) => {
  const { placeholder, renderContent, variant = 'normal' } = props;
  const { results, searchString, setSearchString } = useSearchableListState(props);
  const onSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const searchString = event.target.value;
    setSearchString(searchString);
  };
  const onClear = () => setSearchString('');
  return (
    <>
      <SearchableInput
        onClear={onClear}
        onSearchChange={onSearchChange}
        placeholder={placeholder}
        value={searchString}
        variant={variant}
      />
      {renderContent(results)}
    </>
  );
};

export { SearchResult, PropertyGetter };
