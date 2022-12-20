import { TextField } from '@material-ui/core';
import * as React from 'react';
import t from './strings';
import { InputProps } from './types';
import { getLaunchInputId } from './utils';

/** Shared renderer for any launch input type we can't accept via the UI */
export const NoneInput: React.FC<InputProps> = (props) => {
  const { description, label, name } = props;
  return (
    <TextField
      id={getLaunchInputId(name)}
      fullWidth={true}
      label={label}
      variant="outlined"
      disabled={true}
      helperText={description}
      value={t('noneInputTypeDescription')}
    />
  );
};
