import {
    FormControl,
    FormControlLabel,
    FormHelperText,
    MenuItem,
    Select,
    Switch,
    TextField
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import * as React from 'react';
import { DatetimeInput } from './DatetimeInput';
import { makeStringChangeHandler, makeSwitchChangeHandler } from './handlers';
import { InputProps, InputType } from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { getLaunchInputId } from './utils';

const useStyles = makeStyles(() => ({
    formControl: {
        minWidth: '100%'
    }
}));

/** Handles rendering of the input component for any primitive-type input */
export const SimpleInput: React.FC<InputProps> = props => {
    const {
        error,
        label,
        name,
        onChange,
        typeDefinition: { type, literalType },
        value = ''
    } = props;
    const hasError = !!error;
    const helperText = hasError ? error : props.helperText;
    const classes = useStyles();

    const handleEnumChange = (event: React.ChangeEvent<{ value: unknown }>) => {
        onChange(event.target.value as string);
    };

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
        case InputType.Enum:
            return (
                <FormControl className={classes.formControl}>
                    <Select
                        id={getLaunchInputId(name)}
                        label={label}
                        value={value}
                        onChange={handleEnumChange}
                    >
                        {literalType &&
                            literalType.enumType?.values.map(item => (
                                <MenuItem value={item}>{item}</MenuItem>
                            ))}
                    </Select>
                    <FormHelperText>{label}</FormHelperText>
                </FormControl>
            );
        default:
            return <UnsupportedInput {...props} />;
    }
};
