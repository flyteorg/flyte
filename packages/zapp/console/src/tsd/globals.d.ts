/** is code running in server environment? This value is provided by Webpack */
declare const __isServer: boolean;

/** Represents a plain object where string keys map to values of the same type */
type Dictionary<T> = { [k: string]: T };

/** Shorthand for defining a simple arrow function in a more readable way
 * because `const myFunc: () => string = () => 'the value';` is hard to read.
 * `const myFunc: Fn<string> = () => 'the value'` is a little better.
 */
type Fn<ReturnType> = () => ReturnType;
type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;

type RequireFields<T, FieldTypes extends keyof T> = T & { [K in FieldTypes]-?: NonNullable<T[K]> };
type RequiredNonNullable<T> = RequireFields<T, keyof T>;
