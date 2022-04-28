import { Protobuf } from 'flyteidl';
import * as Long from 'long';

/** Determines if a given date string or object is a valid, usable date. This will detect
 * JS Date objects which have been initialized with invalid values as well as strings which
 * cannot be successfully converted to a valid JS Date.
 */
export function isValidDate(input: string | Date): boolean {
  const date = input instanceof Date ? input : new Date(input);
  const time = date.getTime();
  return !isNaN(time) && time > 0;
}

/** Converts a Protobuf Timestamp object to a JS Date */
export function timestampToDate(timestamp: Protobuf.ITimestamp): Date {
  const nanos = timestamp.nanos || 0;
  const milliseconds = (timestamp.seconds as Long).toNumber() * 1000 + nanos / 1e6;
  return new Date(milliseconds);
}

/** A sort comparison function for ordering timestamps in ascending progression */
export function compareTimestampsAscending(a: Protobuf.ITimestamp, b: Protobuf.ITimestamp) {
  const leftSeconds: Long = a.seconds || Long.fromNumber(0);
  const leftNanos: number = a.nanos || 0;
  const rightSeconds: Long = b.seconds || Long.fromNumber(0);
  const rightNanos: number = b.nanos || 0;
  if (leftSeconds.eq(rightSeconds)) {
    return leftNanos - rightNanos;
  }
  return leftSeconds.lt(rightSeconds) ? -1 : 1;
}

export function dateToTimestamp(date: Date): Protobuf.Timestamp {
  const ms = date.getTime();
  return Protobuf.Timestamp.create({
    seconds: Long.fromNumber(ms / 1000),
    nanos: (ms % 1000) * 1000000,
  });
}

/** Converts a Protobuf Duration object to its equivalent value in milliseconds */
export function durationToMilliseconds(duration: Protobuf.IDuration): number {
  const nanos = duration.nanos || 0;
  return (duration.seconds as Long).toNumber() * 1000 + nanos / 1e6;
}

/** Converts a (possibly fractional) value in milliseconds to a Protobuf Duration object */
export function millisecondsToDuration(milliseconds: number): Protobuf.Duration {
  return new Protobuf.Duration({
    seconds: Long.fromNumber(Math.floor(milliseconds / 1000)),
    // Nanosecond resolution is more than enough, this is to prevent precision errors
    nanos: Math.floor(milliseconds * 1e6) % 1e9,
  });
}

/** Ensures that a string is slash-prefixed */
export function ensureSlashPrefixed(path: string) {
  return path.startsWith('/') ? path : `/${path}`;
}

/** Creates a URL to the same host with a given path */
export function createLocalURL(path: string) {
  return `${window.location.origin}${ensureSlashPrefixed(path)}`;
}

/** Returns entires for an object, sorted lexicographically */
export function sortedObjectEntries<T = unknown>(object: { [s: string]: T }): [string, T][] {
  return Object.entries(object).sort((a, b) => a[0].localeCompare(b[0]));
}

/** Returns keys for an objext, sorted lexicographically */
export function sortedObjectKeys(object: Record<string, unknown>): ReturnType<typeof Object.keys> {
  return Object.keys(object).sort((a, b) => a.localeCompare(b));
}

/**
 * Helper function for exhaustive checks of discriminated unions.
 * https://basarat.gitbooks.io/typescript/docs/types/discriminated-unions.html
 */
export function assertNever(value: never, { noThrow }: { noThrow?: boolean } = {}): never {
  if (noThrow) {
    return value;
  }

  throw new Error(`Unhandled discriminated union member: ${JSON.stringify(value)}`);
}

/**
 * Function that returns boolean value for the string or number representation
 */
export function toBoolean(value?: string): boolean {
  if (value == null) {
    return false;
  }
  return ['true', 'True', 'TRUE', '1'].includes(value);
}

/** Simple shared stringify function to ensure consistency in formatting with
 * respect to spacing.
 */
export function stringifyValue(value: unknown): string {
  return JSON.stringify(value, null, 2);
}
