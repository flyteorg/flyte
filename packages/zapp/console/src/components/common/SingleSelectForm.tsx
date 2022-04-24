import {
  FormControl,
  FormControlLabel,
  FormLabel,
  makeStyles,
  Radio,
  RadioGroup,
  Theme,
} from '@material-ui/core';
import * as React from 'react';
import { useCommonStyles } from './styles';

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    lineHeight: 1.5,
    textTransform: 'uppercase',
    color: theme.palette.text.secondary,
  },
  group: {
    color: theme.palette.text.primary,
  },
}));

export interface SelectValue {
  label: string;
  value: string;
  data: any;
}

export interface SingleSelectFormProps {
  label: string;
  onChange: (newValue: string) => void;
  values: SelectValue[];
  selectedValue: string;
}

/** Form content for rendering a header and list of radios. */
export const SingleSelectForm: React.FC<SingleSelectFormProps> = ({
  label,
  onChange,
  values,
  selectedValue,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const handleChange = (event: React.ChangeEvent<{}>, value: string) => onChange(value);

  return (
    <div>
      <FormControl component="fieldset">
        <FormLabel component="legend" className={styles.title}>
          {label}
        </FormLabel>
        <RadioGroup
          aria-label={label}
          name={label}
          className={styles.group}
          value={selectedValue}
          onChange={handleChange}
        >
          {values.map(({ label, value }) => (
            <FormControlLabel
              className={commonStyles.formControlLabelSmall}
              key={label}
              value={value}
              control={<Radio className={commonStyles.formControlSmall} size="small" />}
              label={label}
            />
          ))}
        </RadioGroup>
      </FormControl>
    </div>
  );
};
