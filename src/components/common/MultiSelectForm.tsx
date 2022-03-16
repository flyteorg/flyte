import {
  Checkbox,
  FormControl,
  FormControlLabel,
  FormGroup,
  FormLabel,
  Link,
  makeStyles,
  Theme,
} from '@material-ui/core';
import * as React from 'react';
import { useCommonStyles } from './styles';

const useStyles = makeStyles((theme: Theme) => ({
  group: {
    color: theme.palette.text.primary,
  },
  listHeader: {
    color: theme.palette.text.secondary,
    lineHeight: 1.5,
    textTransform: 'uppercase',
  },
  resetLink: {
    marginLeft: theme.spacing(4),
    width: theme.spacing(5),
  },
  title: {
    display: 'flex',
    alignItems: 'center',
    margin: 0,
  },
}));

export interface SelectValue {
  label: string;
  value: string;
  data: any;
}

export interface MultiSelectFormProps {
  active?: boolean;
  label: string;
  listHeader?: string;
  onChange: (selectedStates: Record<string, boolean>) => void;
  onReset?: () => void;
  values: SelectValue[];
  selectedStates: Record<string, boolean>;
}

/** Form content for rendering a header and list of checkboxes. If the form is
 * active and a `onReset` handler is provided, a reset link will also be
 * rendered.
 */
export const MultiSelectForm: React.FC<MultiSelectFormProps> = ({
  active = false,
  label,
  listHeader = '',
  onChange,
  onReset,
  values,
  selectedStates,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const showReset = active && onReset;

  const handleChange =
    (name: string) =>
    ({ target: { checked } }: React.ChangeEvent<HTMLInputElement>) =>
      onChange({ ...selectedStates, [name]: checked });

  const handleClickReset = () =>
    onChange(values.reduce((out, { value }) => ({ ...out, [value]: false }), {}));

  // Using a placeholder element with fixed horizontal
  // layout to prevent the form changing size when we show the
  // reset link
  const resetControl = showReset ? (
    <Link
      className={styles.resetLink}
      component="button"
      variant="body1"
      onClick={handleClickReset}
    >
      Reset
    </Link>
  ) : (
    <div className={styles.resetLink} />
  );
  return (
    <div>
      <FormControl aria-label={label} component="fieldset">
        <div className={styles.title}>
          <FormLabel className={styles.listHeader}>{listHeader}</FormLabel>
          {resetControl}
        </div>
        <FormGroup className={styles.group}>
          {values.map(({ label, value }) => {
            const control = (
              <Checkbox
                className={commonStyles.formControlSmall}
                // Casting because the type of the first argument
                // of the handler is React.ChangeEvent<{}> instead
                // of the more specific one for HTMLInputElement
                onChange={handleChange(value) as any}
                checked={!!selectedStates[value]}
                size="small"
                value={value}
              />
            );
            return (
              <FormControlLabel
                className={commonStyles.formControlLabelSmall}
                control={control}
                key={label}
                label={label}
              />
            );
          })}
        </FormGroup>
      </FormControl>
    </div>
  );
};
