import { Core } from 'flyteidl';
import { mapValues } from 'lodash';
import { Scalar } from 'models/Common/types';
import { binaryValues } from './binaryValues';
import { blobValues } from './blobValues';
import { errorValues } from './errorValues';
import { primitiveValues } from './primitiveValues';
import { schemaValues } from './schemaValues';

/** Maps an dictionary of raw values to a dictionary of scalars, each with
 * the correct `value` to allow lookup of whichever field is populated.
 */
function toScalars<T>(type: keyof Core.IScalar, values: Dictionary<T>): Dictionary<Scalar> {
  return mapValues(values, (value: T) => ({
    value: type,
    [type]: value,
  }));
}

export const binaryScalars = toScalars('binary', binaryValues);
export const blobScalars = toScalars('blob', blobValues);
export const errorScalars = toScalars('error', errorValues);
export const primitiveScalars = toScalars('primitive', primitiveValues);
export const schemaScalars = toScalars('schema', schemaValues);

export const noneTypeScalar: Scalar = { value: 'noneType' };
