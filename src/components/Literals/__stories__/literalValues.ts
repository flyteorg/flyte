import { Core } from 'flyteidl';
import { mapValues } from 'lodash';
import { Literal } from 'models/Common/types';
import {
  binaryScalars,
  blobScalars,
  errorScalars,
  noneTypeScalar,
  primitiveScalars,
  schemaScalars,
} from './scalarValues';

/** Maps a dictionary of `Scalar`s to a dictionary of `Literals` by wrapping
 * each value in a `Literal` and setting the appropriate `value` to allow
 * lookup of whichever field is populated.
 */
function toLiterals<T>(type: keyof Core.ILiteral, values: Dictionary<T>): Dictionary<Literal> {
  return mapValues(values, (value: T) => ({
    value: type,
    [type]: value,
  }));
}

export const binaryLiterals = toLiterals('scalar', binaryScalars);
export const blobLiterals = toLiterals('scalar', blobScalars);
export const errorLiterals = toLiterals('scalar', errorScalars);
export const primitiveLiterals = toLiterals('scalar', primitiveScalars);
export const schemaLiterals = toLiterals('scalar', schemaScalars);

export const noneTypeLiteral: Literal = {
  value: 'scalar',
  scalar: noneTypeScalar,
};
