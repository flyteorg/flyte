import { obj } from 'test/utils';

export enum RequestErrorType {
  NotFound = 404,
  Unexpected = 500,
}

export class RequestError extends Error {
  constructor(public type: RequestErrorType, message?: string) {
    super(message);
  }
}

export function notFoundError(id: unknown): RequestError {
  return new RequestError(RequestErrorType.NotFound, `Couldn't find item: ${obj(id)}`);
}

export function unexpectedError(message: string): RequestError {
  return new RequestError(RequestErrorType.Unexpected, message);
}
