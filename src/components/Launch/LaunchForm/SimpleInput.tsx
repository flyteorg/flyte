import {
    FormControl,
    FormControlLabel,
    FormHelperText,
    Switch,
    TextField
} from '@material-ui/core';
import * as React from 'react';
import { DatetimeInput } from './DatetimeInput';
import { makeStringChangeHandler, makeSwitchChangeHandler } from './handlers';
import { InputProps, InputType } from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { getLaunchInputId } from './utils';

/** Handles rendering of the input component for any primitive-type input */
export const SimpleInput: React.FC<InputProps> = props => {
    const {
        error,
        label,
        name,
        onChange,
        typeDefinition: { type },
        value = ''
    } = props;
    const hasError = !!error;
    const helperText = hasError ? error : props.helperText;
    switch (type) {
        case InputType.Boolean:
            return (
                <FormControl>
                    <FormControlLabel
                        control={
                            <Switch
                                id={getLaunchInputId(name)}
                                checked={!!value}
                                onChange={makeSwitchChangeHandler(onChange)}
                                value={name}
                            />
                        }
                        label={label}
                    />
                    <FormHelperText>{helperText}</FormHelperText>
                </FormControl>
            );
        case InputType.Datetime:
            return <DatetimeInput {...props} />;
        case InputType.Schema:
        case InputType.String:
        case InputType.Integer:
        case InputType.Float:
        case InputType.Duration:
            return (
                <TextField
                    error={hasError}
                    id={getLaunchInputId(name)}
                    helperText={helperText}
                    fullWidth={true}
                    label={label}
                    onChange={makeStringChangeHandler(onChange)}
                    value={value}
                    variant="outlined"
                />
            );
        default:
            return <UnsupportedInput {...props} />;
    }
};
