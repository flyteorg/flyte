import { TextField } from '@material-ui/core';
import * as React from 'react';
import { InputProps } from './types';

/** Shared renderer for any launch input type we can't accept via the UI */
export const UnsupportedInput: React.FC<InputProps> = props => {
    const { description, label } = props;
    return (
        <TextField
            fullWidth={true}
            label={label}
            variant="outlined"
            disabled={true}
            helperText={description}
            value="This type is not supported by Flyte Console"
        />
    );
};
