import * as Long from 'long';
import * as LossLessJSON from 'lossless-json';

function losslessReviver(key: string, value: any) {
  if (value && value.isLosslessNumber) {
    // *All* numbers will be converted to LossLessNumber, but we only want
    // to use a Long if it would overflow
    try {
      return value.valueOf();
    } catch {
      return Long.fromString(value.toString());
    }
  }
  return value;
}

export function parseJSON(value: string) {
  return LossLessJSON.parse(value, losslessReviver);
}
