import { AxiosError } from 'axios';
import { NotAuthorizedError, NotFoundError } from 'errors/fetchErrors';

/** Detects special cases for errors returned from Axios and lets others pass through. */
export function transformRequestError(e: Error, path: string) {
  const error = e as AxiosError;
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
  return error;
}
