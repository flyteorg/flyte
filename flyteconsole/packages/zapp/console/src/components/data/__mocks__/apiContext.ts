import { APIContextValue, defaultAPIContextValue } from '../apiContext';

/** Creates a mock version of the APIContext, where all api functions
 * will return a rejected promise by default. Any passed override implementations
 * will be merged into the returned object.
 */
export function mockAPIContextValue(overrides: Partial<APIContextValue>): APIContextValue {
  const emptyImpl = Object.keys(defaultAPIContextValue).reduce<APIContextValue>(
    (out, key) =>
      Object.assign(out, {
        [key]: () => {
          throw new Error(` ${key} not implemented`);
        },
      }),
    {} as APIContextValue,
  );
  return Object.assign(emptyImpl, overrides);
}
