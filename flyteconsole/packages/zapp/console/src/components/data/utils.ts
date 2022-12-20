import { isObject, isPlainObject } from 'lodash';
import { QueryKey } from 'react-query';

export function normalizeQueryKey(queryKey: QueryKey): QueryKey {
  const arrayQueryKey = Array.isArray(queryKey) ? queryKey : [queryKey];
  // for objects with non-default prototypes (such as decoded protobufJS messages),
  // the built-in serialization won't work correctly. So we will convert them
  // to plain objects by spreading ownProperties into a new object.
  const normalizedKey = arrayQueryKey.map((key) => {
    if (isObject(key) && !isPlainObject(key)) {
      return { ...key };
    }
    return key;
  });
  return normalizedKey;
}
