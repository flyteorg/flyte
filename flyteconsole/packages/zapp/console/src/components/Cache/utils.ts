import * as objectHash from 'object-hash';

export type KeyableType = any[] | object | string | symbol;

/** Generic cache key generator. For object, will generate a unique hash.
 * Strings are passed through for convenience.
 */
export function getCacheKey(id: KeyableType): string {
  if (typeof id === 'symbol') {
    return id.toString();
  }
  return typeof id === 'string'
    ? id
    : // We only want to compare own properties, not .__proto__, .constructor, etc.
      objectHash(id, { respectType: false });
}
