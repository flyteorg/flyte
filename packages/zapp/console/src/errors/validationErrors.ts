import { ParameterError } from './parameterErrors';

/** Indicates that one or more provided values are invalid */
export class ValidationError extends Error {
  constructor(public errors: ParameterError[], msg = 'One or more parameters are invalid') {
    super(msg);
  }
}
