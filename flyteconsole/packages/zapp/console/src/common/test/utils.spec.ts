import { long, obj } from 'test/utils';
import { Protobuf } from 'flyteidl';
import {
  compareTimestampsAscending,
  createLocalURL,
  dateToTimestamp,
  durationToMilliseconds,
  ensureSlashPrefixed,
  isValidDate,
  millisecondsToDuration,
  timestampToDate,
} from '../utils';

jest.mock('common/env', () => ({
  env: jest.requireActual('common/env').env,
}));

describe('isValidDate', () => {
  const cases: [string | Date, boolean][] = [
    [new Date(0), false], // This is a valid JS date, but not valid in our system
    ['abc', false],
    [new Date('abc'), false],
    [new Date().toDateString(), true],
    [new Date().toISOString(), true],
    [new Date(), true],
  ];

  cases.forEach(([value, expected]) =>
    it(`should return ${expected} for ${value}`, () => expect(isValidDate(value)).toBe(expected)),
  );
});

describe('timestampToDate', () => {
  const cases: [Protobuf.ITimestamp, Date][] = [
    [{ seconds: long(0), nanos: 0 }, new Date(0)],
    [{ seconds: long(0), nanos: 1e6 }, new Date(1)],
    [{ seconds: long(0), nanos: 1e6 * 30 }, new Date(30)],
    [{ seconds: long(1), nanos: 1e6 * 30 }, new Date(1030)],
  ];

  cases.forEach(([input, expected]) =>
    it(`should return ${expected} for input ${obj(input)}`, () =>
      expect(timestampToDate(input)).toEqual(expected)),
  );
});

describe('dateToTimestamp', () => {
  const cases: [Date, Protobuf.ITimestamp][] = [
    [new Date(0), { seconds: long(0), nanos: 0 }],
    [new Date(1), { seconds: long(0), nanos: 1e6 }],
    [new Date(30), { seconds: long(0), nanos: 1e6 * 30 }],
    [new Date(1030), { seconds: long(1), nanos: 1e6 * 30 }],
  ];

  cases.forEach(([input, expected]) =>
    it(`should return ${obj(expected)} for input ${input} }`, () =>
      expect(dateToTimestamp(input)).toEqual(expected)),
  );
});

describe('durationToMilliseconds', () => {
  const cases: [Protobuf.IDuration, number][] = [
    [{ seconds: long(0), nanos: 0 }, 0],
    [{ seconds: long(0), nanos: 999999 }, 0.999999],
    [{ seconds: long(0), nanos: 1e6 }, 1],
    [{ seconds: long(1), nanos: 0 }, 1000],
    [{ seconds: long(1), nanos: 1e6 }, 1001],
    [{ seconds: long(1), nanos: 1e7 }, 1010],
  ];

  cases.forEach(([input, expected]) =>
    it(`should return ${expected} ms for a duration of ${obj(input)}`, () =>
      expect(durationToMilliseconds(input)).toEqual(expected)),
  );
});

describe('compareTimestampsAscending', () => {
  // Note: Specific cases ignored because they would be invalid timestamps:
  // * A `nanos` value of 1e9 or greater (would roll over into next second)
  const cases: [Protobuf.ITimestamp, Protobuf.ITimestamp, -1 | 0 | 1][] = [
    [{ seconds: long(0), nanos: 0 }, { seconds: long(0), nanos: 1 }, -1],
    [{ seconds: long(0), nanos: 0 }, { seconds: long(0), nanos: 0 }, 0],
    [{ seconds: long(0), nanos: 1 }, { seconds: long(0), nanos: 0 }, 1],
    [{ seconds: long(0), nanos: -1 }, { seconds: long(0), nanos: 0 }, -1],
    [{ seconds: long(0), nanos: 0 }, { seconds: long(0), nanos: -1 }, 1],
    [{ seconds: long(0), nanos: -1 }, { seconds: long(0), nanos: -2 }, 1],
    [{ seconds: long(0), nanos: -2 }, { seconds: long(0), nanos: -1 }, -1],
    [{ seconds: long(-1), nanos: 0 }, { seconds: long(-2), nanos: 0 }, 1],
    [{ seconds: long(-2), nanos: 0 }, { seconds: long(-1), nanos: 0 }, -1],
    [{ seconds: long(10), nanos: 0 }, { seconds: long(9), nanos: 1e9 - 1 }, 1],
    [{ seconds: long(9), nanos: 1e9 - 1 }, { seconds: long(10), nanos: 0 }, -1],
  ];

  cases.forEach(([left, right, expected]) =>
    it(`should return ${expected} when comparing ${obj(left)}, to ${obj(right)}`, () =>
      expect(compareTimestampsAscending(left, right)).toEqual(expected)),
  );
});

describe('millisecondsToDuration', () => {
  const cases: [number, Protobuf.IDuration][] = [
    [0, { seconds: long(0), nanos: 0 }],
    [0.555, { seconds: long(0), nanos: 555000 }],
    [0.999999, { seconds: long(0), nanos: 999999 }],
    // test rounding down to nearest nanosecond
    [0.9999991, { seconds: long(0), nanos: 999999 }],
    [1, { seconds: long(0), nanos: 1e6 }],
    [1000, { seconds: long(1), nanos: 0 }],
    [1001, { seconds: long(1), nanos: 1e6 }],
    [1010, { seconds: long(1), nanos: 1e7 }],
  ];

  cases.forEach(([input, expected]) =>
    it(`should return ${obj(expected)} ms for a duration of ${input}`, () =>
      expect(millisecondsToDuration(input)).toEqual(expected)),
  );
});

describe('ensureSlashPrefixed', () => {
  it('Adds a slash if not present', () => {
    expect(ensureSlashPrefixed('abc')).toBe('/abc');
  });

  it('Does not add slash if already present', () => {
    expect(ensureSlashPrefixed('/abc')).toBe('/abc');
  });
});

describe('createLocalURL', () => {
  it('should append the current origin to the input', () => {
    expect(window.location.origin.length).toBeGreaterThan(0);
    expect(createLocalURL('/test')).toBe(`${window.location.origin}/test`);
  });

  it('should ensure that path is slash-prefixed', () => {
    expect(createLocalURL('test')).toBe(`${window.location.origin}/test`);
  });
});
