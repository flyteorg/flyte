import { Core } from 'flyteidl';

export function primitiveLiteral(primitive: Core.IPrimitive): Core.ILiteral {
    return { scalar: { primitive } };
}
