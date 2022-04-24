class Timer {
  public readonly startTime = window.performance.now();

  public get time() {
    return window.performance.now() - this.startTime;
  }

  public get timeStringMS() {
    return `${this.time.toFixed(2)}ms`;
  }
}

/** Returns a simple object to allow precision timing based on
 * window.performance. On construction, the current timestamp is read.
 * Subsequent calls to the accessor functions will return the time elapsed since
 * `startTime` was sampled.
 */
export function createTimer() {
  return new Timer();
}
