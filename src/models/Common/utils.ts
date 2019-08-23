import { takeWhile } from 'lodash';
import { Identifier } from './types';

/** Concatenates string fields of an Identifier (excluding resourceType) into a
 * predictable output suitable for use as a key.
 */
export function identifierToString({
    domain,
    name,
    project,
    version
}: Identifier): string {
    return `${project}/${domain}/${name}/${version}`;
}

/** Will create a path combining a prefix and a (partial) Identifier. The
 * fields of the Identifier are applied in a specific order: project, domain,
 * name, version. Providing a later value without providing the values that come
 * before it is invalid.
 * ex. { project, domain, name } is valid, { name } is not.
 */
export function makeIdentifierPath(
    prefix: string,
    { project, domain, name, version }: Partial<Identifier>
) {
    const path = takeWhile([project, domain, name, version]).join('/');
    return `${prefix}/${path}`;
}
