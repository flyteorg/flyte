import {
  Button,
  FormControl,
  FormLabel,
  makeStyles,
  OutlinedInput,
  Theme,
} from '@material-ui/core';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    textTransform: 'uppercase',
    color: theme.palette.text.secondary,
  },
  input: {
    margin: `${theme.spacing(1)}px 0`,
  },
}));

export interface SearchInputFormProps {
  label: string;
  placeholder?: string;
  onChange: (newValue: string) => void;
  defaultValue: string;
}

/** Form content for rendering a header and search input. The value is applied
 * on submission of the form.
 */
export const SearchInputForm: React.FC<SearchInputFormProps> = ({
  label,
  placeholder,
  onChange,
  defaultValue,
}) => {
  const [value, setValue] = React.useState(defaultValue);
  const styles = useStyles();
  const onInputChange: React.ChangeEventHandler<HTMLInputElement> = ({ target: { value } }) =>
    setValue(value);

  const onSubmit: React.FormEventHandler = (event) => {
    event.preventDefault();
    onChange(value);
  };
  return (
    <form onSubmit={onSubmit}>
      <FormLabel component="legend" className={styles.title}>
        {label}
      </FormLabel>
      <FormControl margin="dense" variant="outlined" fullWidth={true}>
        <OutlinedInput
          autoFocus={true}
          className={styles.input}
          labelWidth={0}
          onChange={onInputChange}
          placeholder={placeholder}
          type="text"
          value={value}
        />
      </FormControl>
      <Button color="primary" onClick={onSubmit} type="submit" variant="contained">
        Find
      </Button>
    </form>
  );
};
