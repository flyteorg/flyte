import * as Long from 'long';

export const long = (val: number) => Long.fromNumber(val);
export const obj = (val: any) => JSON.stringify(val, null, 2);
export function pendingPromise<T = any>(): Promise<T> {
    return new Promise(() => {});
}
