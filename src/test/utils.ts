import * as Long from 'long';

export const long = (val: number) => Long.fromNumber(val);
export const obj = (val: any) => JSON.stringify(val, null, 2);
export function pendingPromise<T = any>(): Promise<T> {
    return new Promise(() => {});
}

export interface DelayedPromiseResult<T> {
    promise: Promise<T>;
    resolve: (value: T) => void;
    reject: (value: any) => void;
}
export function delayedPromise<T = any>(): DelayedPromiseResult<T> {
    let resolve: (value: any) => void;
    let reject: (value: any) => void;
    const promise = new Promise<T>((innerResolve, innerReject) => {
        resolve = innerResolve;
        reject = innerReject;
    });
    return {
        promise,
        resolve: value => resolve(value),
        reject: value => reject(value)
    };
}

/** Returns a promise that will resolve after the given time. Useful in
 * combination with `await` to pause testing for a short amount of time
 */
export function waitFor(timeMS: number) {
    return new Promise(resolve => {
        setTimeout(resolve, timeMS);
    });
}
