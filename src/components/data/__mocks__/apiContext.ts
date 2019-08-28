import { APIContextValue, defaultAPIContextValue } from '../apiContext';

export function mockAPIContextValue(
    overrides: Partial<APIContextValue>
): APIContextValue {
    const emptyImpl = Object.keys(defaultAPIContextValue).reduce<
        APIContextValue
    >(
        (out, key) =>
            Object.assign(out, {
                [key]: () => Promise.reject(` ${key} not implemented`)
            }),
        {} as APIContextValue
    );
    return Object.assign(emptyImpl, overrides);
}
