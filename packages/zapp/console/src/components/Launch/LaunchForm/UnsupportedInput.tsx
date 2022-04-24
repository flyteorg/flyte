import { TextField } from '@material-ui/core';
import * as React from 'react';
import { InputProps } from './types';
import { getLaunchInputId } from './utils';

/** Shared renderer for any launch input type we can't accept via the UI */
export const UnsupportedInput: React.FC<InputProps> = (props) => {
  const { description, label, name } = props;
  return (
    <TextField
      id={getLaunchInputId(name)}
      fullWidth={true}
      label={label}
      variant="outlined"
      disabled={true}
      helperText={description}
      value="Flyte Console does not support entering values of this type"
    />
  );
};
