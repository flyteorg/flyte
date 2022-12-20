import { TextField } from '@material-ui/core';
import * as React from 'react';
import { log } from 'common/log';
import { makeStringChangeHandler } from './handlers';
import { InputProps, InputType } from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { getLaunchInputId } from './utils';

/** Handles rendering of the input component for a Collection of SimpleType values */
export const CollectionInput: React.FC<InputProps> = (props) => {
  const {
    error,
    label,
    name,
    onChange,
    typeDefinition: { subtype },
    value = '',
  } = props;
  if (!subtype) {
    log.warn('Unexpected missing subtype for collection input', props.typeDefinition);
    return <UnsupportedInput {...props} />;
  }
  const hasError = !!error;
  const helperText = hasError ? error : props.helperText;
  switch (subtype.type) {
    case InputType.Blob:
    case InputType.Boolean:
    case InputType.Collection:
    case InputType.Datetime:
    case InputType.Duration:
    case InputType.Error:
    case InputType.Float:
    case InputType.Integer:
    case InputType.Map:
    case InputType.String:
    case InputType.Struct:
      return (
        <TextField
          id={getLaunchInputId(name)}
          error={hasError}
          helperText={helperText}
          fullWidth={true}
          label={label}
          multiline={true}
          onChange={makeStringChangeHandler(onChange)}
          rowsMax={8}
          value={value}
          variant="outlined"
        />
      );
    default:
      return <UnsupportedInput {...props} />;
  }
};
