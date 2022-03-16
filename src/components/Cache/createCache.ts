import { getCacheKey } from './utils';

type EntityKey = object | string | symbol;
interface HasIdObject {
  id: object;
}

function hasId(value: Object): value is HasIdObject {
  return {}.hasOwnProperty.call(value, 'id');
}

/** A generic cache for any type of object or Array. Keys can be objects,
 * strings, or Symbols. Object keys will be hashed, so they do not need to be
 * strictly equal (===). The cache provides methods for getting/setting values
 * by key and for merging values or arrays of values with any existing items.
 */
export interface ValueCache<ValueType = object> {
  /** Retrieve an item by key */
  get(key: EntityKey): ValueType | undefined;
  /** Check existence of an item by key */
  has(key: EntityKey): boolean;
  /** Merge an array of values. If the items have an `id` property, its value
   * will be hashed to create the key. Otherwise, the item itself will be
   * hashed.
   * @returns A new array containing the merged values
   */
  mergeArray<T extends Object>(values: T[]): T[];
  /** Merge a single value by key. Equivalent to `set()` if no value exists.
   * For values that are objects, a shallow merge with `Object.assign` is
   * performed. For arrays, the value is _replaced_.
   * @returns The merged value
   */
  mergeValue(key: EntityKey, value: ValueType): ValueType;
  /** Set an item value by key. Replaces any existing value. */
  set(key: EntityKey, value: ValueType): ValueType;
}

type Cacheable = object | any[];

/** Creates a new `ValueCache` */
export function createCache(): ValueCache {
  const entities = new Map<string, Cacheable>();

  const get = (id: EntityKey) => {
    return entities.get(getCacheKey(id));
  };

  const has = (id: EntityKey) => {
    return entities.has(getCacheKey(id));
  };

  const set = (id: EntityKey, value: Cacheable) => {
    entities.set(getCacheKey(id), value);
    return value;
  };

  const mergeValue = (id: EntityKey, value: Cacheable) => {
    const key = getCacheKey(id);
    const existing = entities.get(key);

    if (Array.isArray(value)) {
      entities.set(key, value);
      return value;
    }

    const merged = Object.assign({}, existing, value);
    entities.set(key, merged);
    return merged;
  };

  const mergeArray = <T extends Object>(values: T[]) => {
    return values.map<T>((item: T) => {
      const id = hasId(item) ? item.id : item;
      return mergeValue(id, item) as T;
    });
  };

  return { get, has, mergeArray, mergeValue, set };
}
