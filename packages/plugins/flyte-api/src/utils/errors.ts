/* eslint-disable max-classes-per-file */
import { AxiosError } from 'axios';

export class NotFoundError extends Error {
  constructor(public override name: string, msg = 'The requested item could not be found') {
    super(msg);
  }
}

/** Indicates failure to fetch a resource because the user is not authorized (401) */
export class NotAuthorizedError extends Error {
  constructor(msg = 'User is not authorized to view this resource') {
    super(msg);
  }
}

/** Detects special cases for errors returned from Axios and lets others pass through. */
export function transformRequestError(err: unknown, path: string) {
  const error = err as AxiosError;

  if (!error.response) {
    return error;
  }

  // For some status codes, we'll throw a special error to allow
  // client code and components to handle separately
  if (error.response.status === 404) {
    return new NotFoundError(path);
  }
  if (error.response.status === 401) {
    return new NotAuthorizedError();
  }

  // this error is not decoded.
  return error;
}
