import { default as momentDateUtils } from '@date-io/moment'; // choose your lib
import {
    KeyboardDateTimePicker,
    MuiPickersUtilsProvider
} from '@material-ui/pickers';
// Flyte dates are specified in UTC
import { Moment, utc as moment } from 'moment';
import * as React from 'react';
import { InputProps } from './types';

function defaultDate() {
    return moment()
        .startOf('day')
        .toISOString();
}

/** A form field for selecting a date/time from a picker or entering it via
 * keyboard.
 */
export const DatetimeInput: React.FC<InputProps> = props => {
    const { label, helperText, onChange, value: propValue } = props;
    const value = typeof propValue === 'string' ? propValue : defaultDate();

    const handleChange = (
        dateValue: Moment | null,
        stringValue?: string | null
    ) => {
        if (stringValue != null) {
            onChange(stringValue);
        } else if (dateValue !== null) {
            onChange(dateValue.toISOString());
        } else {
            onChange('');
        }
    };
    return (
        <MuiPickersUtilsProvider libInstance={moment} utils={momentDateUtils}>
            <KeyboardDateTimePicker
                ampm={false}
                format="MM/DD/YYYY HH:mm:ss"
                helperText={helperText}
                inputVariant="outlined"
                label={label}
                onChange={handleChange}
                onError={console.log}
                showTodayButton={true}
                value={value}
            />
        </MuiPickersUtilsProvider>
    );
};
