import { Core } from 'flyteidl';
import { ISO_8601, RFC_2822 } from 'moment';

export function literalNone(): Core.ILiteral {
    return { scalar: { noneType: {} } };
}

export const allowedDateFormats = [ISO_8601, RFC_2822];
