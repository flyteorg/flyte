import { millisecondsToDuration } from 'common/utils';
import { Admin } from 'flyteidl';
import * as moment from 'moment-timezone';
import { unknownValueString, zeroSecondsString } from '../constants';
import {
  dateDiffString,
  dateFromNow,
  dateWithFromNow,
  durationToYMWDHMS,
  ensureUrlWithProtocol,
  formatDate,
  formatDateLocalTimezone,
  formatDateUTC,
  getScheduleFrequencyString,
  getScheduleOffsetString,
  leftPaddedNumber,
  millisecondsToHMS,
  protobufDurationToHMS,
} from '../formatters';

jest.mock('../timezone.ts', () => ({
  timezone: 'America/Los_Angeles',
}));

const invalidDates = ['abc', -200, 0];
// Matches strings in the form 01/01/2000 01:01:00 PM  (5 minutes ago)
const dateWithAgoRegex = /^[\w/:\s]+ (AM|PM)\s+UTC\s+\([a\d] (minute|hour|day|second)s? ago\)$/;
const dateFromNowRegex = /^[a\d] (minute|hour|day|second)s? ago$/;
const dateRegex = /^[\w/:\s]+ (AM|PM)/;
const utcDateRegex = /^[\w/:\s]+ (AM|PM) UTC/;
const localDateRegex = /^[\w/:\s]+ (AM|PM) (PDT|PST)/;

describe('dateWithFromNow', () => {
  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date: ${v}`, () => {
      expect(dateWithFromNow(new Date(v))).toEqual(unknownValueString);
    }),
  );

  // Not testing this extensively because it's relying on moment, which is well-tested
  it('Returns a reasonable date string with (ago) text for valid inputs', () => {
    const date = new Date();
    expect(dateWithFromNow(new Date(date.getTime() - 65000))).toMatch(dateWithAgoRegex);
  });
});

describe('dateFromNow', () => {
  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date: ${v}`, () => {
      expect(dateFromNow(new Date(v))).toEqual(unknownValueString);
    }),
  );

  // Not testing this extensively because it's relying on moment, which is well-tested
  it('Returns a reasonable string for valid inputs', () => {
    const date = new Date();
    expect(dateFromNow(new Date(date.getTime() - 125000))).toMatch(dateFromNowRegex);
  });
});

describe('formatDate', () => {
  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date: ${v}`, () => {
      expect(formatDate(new Date(v))).toEqual(unknownValueString);
    }),
  );

  it('returns a reasonable date string for valid inputs', () => {
    expect(formatDate(new Date())).toMatch(dateRegex);
  });
});

describe('formatDateUTC', () => {
  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date: ${v}`, () => {
      expect(formatDateUTC(new Date(v))).toEqual(unknownValueString);
    }),
  );

  it('returns a reasonable date string for valid inputs', () => {
    expect(formatDateUTC(new Date())).toMatch(utcDateRegex);
  });
});

describe('formatDateLocalTimezone', () => {
  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date: ${v}`, () => {
      expect(formatDateLocalTimezone(new Date(v))).toEqual(unknownValueString);
    }),
  );

  it('returns a reasonable date string for valid inputs', () => {
    expect(formatDateLocalTimezone(new Date())).toMatch(localDateRegex);
  });
});

const millisecondToHMSTestCases: [number, string][] = [
  [-1, unknownValueString],
  [0, zeroSecondsString],
  [1, '1ms'],
  [999, '999ms'],
  [1000, '1s'],
  [60000, '1m'],
  [61000, '1m 1s'],
  [60 * 60000, '1h'],
  [60 * 60000 + 1000, '1h 1s'],
  [60 * 60000 + 60000, '1h 1m'],
  [60 * 60000 + 61000, '1h 1m 1s'],
  // For values greater than a day, we just use the hour value
  [24 * 60 * 60000, '24h'],
  [24 * 60 * 60000 + 61000, '24h 1m 1s'],
];

describe('dateDiffString', () => {
  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date on left side: ${v}`, () => {
      expect(dateDiffString(new Date(v), new Date())).toEqual(unknownValueString);
    }),
  );

  invalidDates.forEach((v) =>
    it(`returns a constant string for invalid date on right side: ${v}`, () => {
      expect(dateDiffString(new Date(), new Date(v))).toEqual(unknownValueString);
    }),
  );

  millisecondToHMSTestCases.forEach(([offset, expected]) =>
    it(`should return ${expected} for an offset of ${offset}`, () => {
      const now = new Date();
      const later = new Date(now.getTime() + offset);
      expect(dateDiffString(now, later)).toEqual(expected);
    }),
  );
});

describe('protobufDurationToHMS', () => {
  millisecondToHMSTestCases.forEach(([ms, expected]) =>
    it(`should convert ${ms}ms to ${expected}`, () => {
      expect(protobufDurationToHMS(millisecondsToDuration(ms))).toBe(expected);
    }),
  );
});

describe('millisecondsToHMS', () => {
  millisecondToHMSTestCases.forEach(([ms, expected]) =>
    it(`should convert ${ms}ms to ${expected}`, () => {
      expect(millisecondsToHMS(ms)).toBe(expected);
    }),
  );
});

describe('durationToYMWDHMS', () => {
  // input and expected result
  const cases: [string, string][] = [
    ['P1Y1M1W1D', '(+) 1y 1M 8d'],
    ['P1Y1M1W1DT1H1M1S', '(+) 1y 1M 8d 1h 1m 1s'],
    ['P1Y1M1DT1H1M1S', '(+) 1y 1M 1d 1h 1m 1s'],
    ['P1M1DT1H1M1S', '(+) 1M 1d 1h 1m 1s'],
    ['P1DT1H1M1S', '(+) 1d 1h 1m 1s'],
    ['PT1H1M1S', '(+) 1h 1m 1s'],
    ['PT1M1S', '(+) 1m 1s'],
    ['PT1S', '(+) 1s'],
    ['PT1M-1S', '(+) 59s'],
    ['-P1Y1M1W1D', '(-) 1y 1M 8d'],
    ['-P1Y1M1W1DT1H1M1S', '(-) 1y 1M 8d 1h 1m 1s'],
    ['-P1Y1M1DT1H1M1S', '(-) 1y 1M 1d 1h 1m 1s'],
    ['-P1M1DT1H1M1S', '(-) 1M 1d 1h 1m 1s'],
    ['-P1DT1H1M1S', '(-) 1d 1h 1m 1s'],
    ['-PT1H1M1S', '(-) 1h 1m 1s'],
    ['-PT1M1S', '(-) 1m 1s'],
    ['-PT1S', '(-) 1s'],
    ['PT-1M1S', '(-) 59s'],
    ['', ''],
  ];
  cases.forEach(([input, expected]) =>
    it(`should produce ${expected} with input ${input}`, () => {
      expect(durationToYMWDHMS(moment.duration(input))).toEqual(expected);
    }),
  );
});

describe('getScheduleFrequencyString', () => {
  // input and expected result
  const cases: [Admin.ISchedule, string][] = [
    [{ cronExpression: '* * * * *' }, 'Every minute'],
    [{ cronExpression: '0 20 ? * 3 *' }, 'At 08:00 PM, only on Tuesday'],
    [{ rate: { value: 1, unit: Admin.FixedRateUnit.MINUTE } }, 'Every 1 minutes'],
    [{ cronSchedule: { schedule: '* * * * *' } }, 'Every minute'],
    [{ cronSchedule: { schedule: '@hourly' } }, 'Every hour'],
    [{ cronSchedule: { schedule: 'hourly' } }, 'Every hour'],
    [{ cronSchedule: { schedule: 'hours' } }, 'Every hour'],
    [{ cronSchedule: { schedule: '@daily' } }, 'Every day'],
    [{ cronSchedule: { schedule: 'daily' } }, 'Every day'],
    [{ cronSchedule: { schedule: 'days' } }, 'Every day'],
    [{ cronSchedule: { schedule: '@weekly' } }, 'Every week'],
    [{ cronSchedule: { schedule: 'weekly' } }, 'Every week'],
    [{ cronSchedule: { schedule: 'weeks' } }, 'Every week'],
    [{ cronSchedule: { schedule: '@monthly' } }, 'Every month'],
    [{ cronSchedule: { schedule: 'monthly' } }, 'Every month'],
    [{ cronSchedule: { schedule: 'months' } }, 'Every month'],
    [{ cronSchedule: { schedule: '@yearly' } }, 'Every year'],
    [{ cronSchedule: { schedule: 'yearly' } }, 'Every year'],
    [{ cronSchedule: { schedule: 'years' } }, 'Every year'],
    [{ cronSchedule: { schedule: '@annually' } }, 'Every year'],
    [{ cronSchedule: { schedule: 'annually' } }, 'Every year'],
    [null!, ''],
    [{ cronSchedule: { schedule: '' } }, ''],
  ];

  cases.forEach(([input, expected]) =>
    it(`should produce ${expected} with input ${input}`, () => {
      expect(getScheduleFrequencyString(input)).toEqual(expected);
    }),
  );
});

describe('getScheduleOffsetString', () => {
  // input and expected result
  const cases: [Admin.ISchedule, string][] = [
    [{ cronSchedule: { offset: 'P1D' } }, '(+) 1d'],
    [{ cronSchedule: { offset: 'P-1D' } }, '(-) 1d'],
    [null!, ''],
    [{ cronSchedule: { offset: '' } }, ''],
  ];

  cases.forEach(([input, expected]) =>
    it(`should produce ${expected} with input ${input}`, () => {
      expect(getScheduleOffsetString(input)).toEqual(expected);
    }),
  );
});

describe('ensureUrlWithProtocol', () => {
  // input and expected result
  const cases: [string, string][] = [
    ['localhost', 'https://localhost'],
    ['http://localhost', 'http://localhost'],
    ['https://localhost', 'https://localhost'],
    // There could be more test cases, but this function is only designed
    // to add a protocol if missing and preserve http:// if it is used
  ];

  cases.forEach(([input, expected]) =>
    it(`should produce ${expected} with input ${input}`, () => {
      expect(ensureUrlWithProtocol(input)).toEqual(expected);
    }),
  );
});

describe('leftPaddedNumber', () => {
  const cases: [number, number, string][] = [
    [1, 0, '1'],
    [10, 0, '10'],
    [0, 1, '0'],
    [0, 2, '00'],
    [1, 1, '1'],
    [1, 2, '01'],
    [1, 3, '001'],
    [10, 1, '10'],
    [10, 2, '10'],
    [10, 3, '010'],
  ];

  cases.forEach(([value, width, expected]) =>
    it(`should produce ${expected} with input (${value}, ${width})`, () => {
      expect(leftPaddedNumber(value, width)).toEqual(expected);
    }),
  );
});
