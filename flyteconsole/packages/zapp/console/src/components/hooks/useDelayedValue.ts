import { useEffect, useState } from 'react';

/** Returns an initial value until `delayMS` has passed and then returns the final value.
 * Useful for delaying an action on mount for a specific amount of time.
 */
export function useDelayedValue<T>(initialValue: T, delayMS: number, finalValue: T) {
  const [result, setValue] = useState<T>(initialValue);
  useEffect(() => {
    const timerId = setTimeout(() => {
      setValue(finalValue);
    }, delayMS);
    return () => clearTimeout(timerId);
  }, [delayMS, finalValue]);
  return result;
}
