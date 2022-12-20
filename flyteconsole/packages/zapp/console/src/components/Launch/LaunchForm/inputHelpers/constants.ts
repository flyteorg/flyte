import { Core } from 'flyteidl';
import { ISO_8601, RFC_2822 } from 'moment';

export function literalNone(): Core.ILiteral {
  return { scalar: { noneType: {} } };
}

export const allowedDateFormats = [ISO_8601, RFC_2822];

const primitivePath = 'scalar.primitive';
export const schemaUriPath = 'scalar.schema.uri';
export const structPath = 'scalar.generic';

/** Strings constants which can be used to perform a deep `get` on a scalar
 * literal type using a primitive value.
 */
export const primitiveLiteralPaths = {
  scalarBoolean: `${primitivePath}.boolean`,
  scalarDatetime: `${primitivePath}.datetime`,
  scalarDuration: `${primitivePath}.duration`,
  scalarFloat: `${primitivePath}.floatValue`,
  scalarInteger: `${primitivePath}.integer`,
  scalarString: `${primitivePath}.stringValue`,
};
