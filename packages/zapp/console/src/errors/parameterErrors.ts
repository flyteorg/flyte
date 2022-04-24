/* eslint-disable max-classes-per-file */

/** Indicates a generic problem with a function parameter */
export class ParameterError extends Error {
  constructor(public name: string, msg: string) {
    super(msg);
  }
}

/** Indicates that a parameter requires a value and none was provided */
export class RequiredError extends ParameterError {
  constructor(public name: string, msg = 'This value is required') {
    super(name, msg);
  }
}

/** Indicates that the provided parameter value is invalid */
export class ValueError extends ParameterError {
  constructor(public name: string, msg = 'Invalid value') {
    super(name, msg);
  }
}
