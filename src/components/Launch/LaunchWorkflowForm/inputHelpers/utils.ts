import { Core } from 'flyteidl';
import { get } from 'lodash';
import { InputType } from '../types';

/** Performs a deep get of `path` on the given `Core.ILiteral`. Will throw
 * if the given property doesn't exist.
 */
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

/** Converts a value within a collection to the appropriate string
 * representation. Some values require additional quotes.
 */
export function collectionChildToString(type: InputType, value: any) {
    if (value === undefined) {
        return '';
    }
    return type === InputType.Integer ? `${value}` : JSON.stringify(value);
}
