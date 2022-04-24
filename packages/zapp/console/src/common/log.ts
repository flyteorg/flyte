import debug from 'debug';

// For now, logging will just go to the window
export const log = window.console;
export const debugPrefix = 'flyte';

export const createDebugLogger = (namespace: string) => debug(`${debugPrefix}:${namespace}`);
export { debug };
