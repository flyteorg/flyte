import { useEffect, useRef, useState } from 'react';

/** Debounces changes to a value using the given delay. Useful, for instance
 * with API calls made by a search/autocomplete input.
 */
export function useDebouncedValue<T>(value: T, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);
  const hasRunOnce = useRef<boolean>(false);

  useEffect(() => {
    if (!hasRunOnce.current) {
      hasRunOnce.current = true;
      return undefined;
    }
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}
