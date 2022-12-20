import { Protobuf } from 'flyteidl';
import { isObject, isPlainObject } from 'lodash';

/** Offsets a given `Protobuf.ITimestamp` by a value in seconds. Useful
 * for creating start time relationships between parent/child or sibling executions.
 */
export function timeStampOffset(
  timeStamp: Protobuf.ITimestamp,
  offsetSeconds: number,
): Protobuf.Timestamp {
  const output = new Protobuf.Timestamp(timeStamp);
  output.seconds =
    offsetSeconds < 0 ? output.seconds.subtract(offsetSeconds) : output.seconds.add(offsetSeconds);
  return output;
}

function stableStringifyReplacer(_key: string, value: unknown): unknown {
  if (typeof value === 'function') {
    throw new Error('Encountered function() during serialization');
  }

  if (isObject(value)) {
    const plainObject: any = isPlainObject(value) ? value : { ...value };
    return Object.keys(plainObject)
      .sort()
      .reduce((result, key) => {
        result[key] = plainObject[key];
        return result;
      }, {} as any);
  }

  return value;
}

/** A copy of the hash function from react-query, useful for generating a unique
 * key for mapping a Request to its associated data based on URL/type/query params/etc.
 */
export function stableStringify(value: unknown): string {
  return JSON.stringify(value, stableStringifyReplacer);
}
