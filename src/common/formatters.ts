import cronstrue from 'cronstrue';
import * as moment from 'moment';

import { Admin, Protobuf } from 'flyteidl';
import { unknownValueString } from './constants';
import { durationToMilliseconds, isValidDate } from './utils';

/** Formats a date into a standard string with a moment-style "from now" hint
 * ex. 12/21/2017 8:19:36 PM (18 days ago)
 */
export function dateWithFromNow(input: Date) {
    if (!isValidDate(input)) {
        return unknownValueString;
    }

    const date = moment.utc(input);
    return `${date.format('l  LTS')} UTC (${date.fromNow()})`;
}

/** Formats a date into a moment-style "from now" value */
export function dateFromNow(input: Date) {
    if (!isValidDate(input)) {
        return unknownValueString;
    }

    const date = moment(input);
    return `${date.fromNow()}`;
}

/** Formats a date into a standard format used throughout the UI
 * ex 12/21/2017 8:19:36 PM
 */
export function formatDate(input: Date) {
    return isValidDate(input)
        ? moment(input).format('l LTS')
        : unknownValueString;
}

/** Formats a date into a standard format used throughout the UI
 * ex 12/21/2017 8:19:36 PM
 */
export function formatDateUTC(input: Date) {
    return isValidDate(input)
        ? `${moment.utc(input).format('l LTS')} UTC`
        : unknownValueString;
}

/** Outputs a value in milliseconds in (H M S) format (ex. 2h 3m 30s) */
export function millisecondsToHMS(valueMS: number): string {
    const duration = moment.duration(valueMS);
    const parts = [];

    // Using asHours() because if it's greater than 24, we'll just show the total
    if (duration.asHours() >= 1) {
        parts.push(`${Math.floor(duration.asHours())}h`);
    }

    if (duration.minutes() >= 1) {
        parts.push(`${duration.minutes()}m`);
    }

    if (duration.seconds() >= 1) {
        parts.push(`${duration.seconds()}s`);
    }

    return parts.length ? parts.join(' ') : unknownValueString;
}

/** Converts a protobuf Duration value to (H M S) format (ex. 2h 3m 30s)*/
export function protobufDurationToHMS(duration: Protobuf.IDuration) {
    return millisecondsToHMS(durationToMilliseconds(duration));
}

/** Calculates the difference between two Dates and outputs it in (H M S) format (ex. 2h 3m 30s)
 */
export function dateDiffString(fromDate: Date, toDate: Date) {
    if (!isValidDate(fromDate) || !isValidDate(toDate)) {
        return unknownValueString;
    }

    return millisecondsToHMS(moment(toDate).diff(fromDate));
}

const fixedRateUnitStrings = {
    [Admin.FixedRateUnit.DAY]: 'days',
    [Admin.FixedRateUnit.HOUR]: 'hours',
    [Admin.FixedRateUnit.MINUTE]: 'minutes'
};

/** Converts a IFixedRate value into a human-readable string ('Every x minutes/hours/days') */
export function fixedRateToString({ value, unit }: Admin.IFixedRate): string {
    if (unit == null || !(unit in Admin.FixedRateUnit) || !value) {
        return '';
    }
    return `Every ${value} ${fixedRateUnitStrings[unit]}`;
}

export function getScheduleFrequencyString(schedule?: Admin.ISchedule) {
    if (schedule == null) {
        return '';
    }
    if (schedule.cronExpression) {
        // Need to add a leading 0 to get a valid CRON expression, because
        // ISchedule is using AWS-style expressions, which don't allow a `seconds` position
        return cronstrue.toString(`0 ${schedule.cronExpression}`);
    }
    if (schedule.rate) {
        return fixedRateToString(schedule.rate);
    }
    return '';
}

/** Ensures that a given string has a protocol prefix. If not, will add https:// to the beginning  */
export function ensureUrlWithProtocol(url: string): string {
    if (url.indexOf('http') !== 0) {
        return `https://${url}`;
    }
    return url;
}
