import * as objectHash from 'object-hash';

/** Generic cache key generator. For object, will generate a unique hash.
 * Strings are passed through for convenience.
 */
export function getCacheKey(id: any[] | object | string) {
    return typeof id === 'string' || typeof id === 'symbol'
        ? id
        : // We only want to compare own properties, not .__proto__, .constructor, etc.
          objectHash(id, { respectType: false });
}
