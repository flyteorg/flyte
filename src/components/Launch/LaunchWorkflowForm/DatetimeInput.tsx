import { default as momentDateUtils } from '@date-io/moment'; // choose your lib
import {
    KeyboardDateTimePicker,
    MuiPickersUtilsProvider
} from '@material-ui/pickers';
// Flyte dates are specified in UTC
import { Moment, utc as moment } from 'moment';
import * as React from 'react';
import { InputProps } from './types';

/** A form field for selecting a date/time from a picker or entering it via
 * keyboard.
 */
export const DatetimeInput: React.FC<InputProps> = props => {
    const { error, label, onChange, value: propValue } = props;
    const hasError = !!error;
    const helperText = hasError ? error : props.helperText;
    const value =
        typeof propValue === 'string' && propValue.length > 0
            ? propValue
            : null;

    const handleChange = (
        dateValue: Moment | null,
        stringValue?: string | null
    ) => {
        if (dateValue && dateValue.isValid()) {
            onChange(dateValue.toISOString());
        } else if (stringValue != null) {
            onChange(stringValue);
        } else {
            onChange('');
        }
    };
    return (
        <MuiPickersUtilsProvider libInstance={moment} utils={momentDateUtils}>
            <KeyboardDateTimePicker
                error={hasError}
                clearable={true}
                fullWidth={true}
                ampm={false}
                format="MM/DD/YYYY HH:mm:ss"
                helperText={helperText}
                inputVariant="outlined"
                label={label}
                onChange={handleChange}
                showTodayButton={true}
                value={value}
            />
        </MuiPickersUtilsProvider>
    );
};
