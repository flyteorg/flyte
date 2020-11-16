import { TextField } from '@material-ui/core';
import * as React from 'react';
import { makeStringChangeHandler } from './handlers';
import { InputProps } from './types';
import { getLaunchInputId } from './utils';

/** Handles rendering of the input component for a Struct */
export const StructInput: React.FC<InputProps> = props => {
    const {
        error,
        label,
        name,
        onChange,
        typeDefinition: { subtype },
        value = ''
    } = props;
    const hasError = !!error;
    const helperText = hasError ? error : props.helperText;
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
};
