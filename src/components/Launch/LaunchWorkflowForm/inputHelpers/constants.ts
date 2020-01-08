import { Core } from 'flyteidl';
import { ISO_8601, RFC_2822 } from 'moment';

export function literalNone(): Core.ILiteral {
    return { scalar: { noneType: {} } };
}

export const allowedDateFormats = [ISO_8601, RFC_2822];

const primitivePath = 'scalar.primitive';

export const literalValuePaths = {
    scalarBoolean: `${primitivePath}.boolean`,
    scalarDatetime: `${primitivePath}.datetime`,
    scalarDuration: `${primitivePath}.duration`,
    scalarFloat: `${primitivePath}.floatValue`,
    scalarInteger: `${primitivePath}.integer`,
    scalarString: `${primitivePath}.stringValue`
};
