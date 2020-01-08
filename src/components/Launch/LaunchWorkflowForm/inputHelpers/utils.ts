import { get } from 'lodash';
import { Literal } from 'models';

export function extractLiteralWithCheck<T>(literal: Literal, path: string): T {
    const value = get(literal, path);
    if (value === undefined) {
        throw new Error(`Failed to extract literal value with path ${path}`);
    }
    return value as T;
}
