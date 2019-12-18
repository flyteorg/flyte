import { Core } from 'flyteidl';

export function literalNone(): Core.ILiteral {
    return { scalar: { noneType: {} } };
}
