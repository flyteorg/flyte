import { Core } from 'flyteidl';
import { get } from 'lodash';

export function extractLiteralWithCheck<T>(
    literal: Core.ILiteral,
    path: string
): T {
    const value = get(literal, path);
    if (value === undefined) {
        throw new Error(`Failed to extract literal value with path ${path}`);
    }
    return value as T;
}
