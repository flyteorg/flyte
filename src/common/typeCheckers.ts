/** Returns if value is a function, definitively typing it if so. */
export function isFunction(value: unknown): value is Function {
  return typeof value === 'function';
}
