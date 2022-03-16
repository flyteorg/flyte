import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import * as Long from 'long';
import { Primitive } from 'models/Common/types';

export const primitiveValues: Dictionary<Partial<Primitive>> = {
  booleanTrue: {
    value: 'boolean',
    boolean: true,
  },
  booleanFalse: {
    value: 'boolean',
    boolean: false,
  },
  integerSmall: {
    value: 'integer',
    integer: Long.fromNumber(10),
  },
  integerLarge: {
    value: 'integer',
    integer: Long.fromNumber(Number.MAX_SAFE_INTEGER).add(5),
  },
  dateTime: {
    value: 'datetime',
    datetime: dateToTimestamp(new Date()),
  },
  duration: {
    value: 'duration',
    duration: millisecondsToDuration(3725000),
  },
  aString: {
    value: 'stringValue',
    stringValue: 'some random string value',
  },
  aFloat: {
    value: 'floatValue',
    floatValue: 1234.5678,
  },
};
