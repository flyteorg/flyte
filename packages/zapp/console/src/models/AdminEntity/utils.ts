import { env } from 'common/env';
import { createDebugLogger } from 'common/log';
import { createLocalURL, ensureSlashPrefixed } from 'common/utils';
import { apiPrefix } from './constants';
import {
  AdminEntityTransformer,
  DecodableType,
  EncodableType,
  PaginatedEntityResponse,
} from './types';

const debug = createDebugLogger('adminEntity');
const loginEndpoint = '/login';
const profileEndpoint = '/me';
const redirectParam = 'redirect_url';

/** Converts a path into a full Admin API url */
export function adminApiUrl(url: string) {
  const finalUrl = ensureSlashPrefixed(url);
  if (env.ADMIN_API_URL) {
    return `${env.ADMIN_API_URL}${apiPrefix}${finalUrl}`;
  }
  return createLocalURL(`${apiPrefix}${finalUrl}`);
}

/** Constructs a url for redirecting to the Admin login endpoint and returning
 * to the current location after completing the flow.
 */
export function getLoginUrl(redirectUrl: string = window.location.href) {
  const baseUrl = env.ADMIN_API_URL
    ? `${env.ADMIN_API_URL}${loginEndpoint}`
    : createLocalURL(loginEndpoint);
  return `${baseUrl}?${redirectParam}=${redirectUrl}`;
}

/** Constructs a URL for fetching the current user profile. */
export function getProfileUrl() {
  if (env.ADMIN_API_URL) {
    return `${env.ADMIN_API_URL}${profileEndpoint}`;
  }
  return createLocalURL(profileEndpoint);
}

// Helper to log out the contents of a protobuf response, since the Network tab
// shows binary values :-).
export function logProtoResponse<T>(url: string, data: T): T {
  debug(`Request: ${url}, \n%O`, data);
  return data;
}

export function decodeProtoResponse<T>(data: ArrayBuffer, messageType: DecodableType<T>): T {
  // ProtobufJS requires Uint8Array, but axios returns an ArrayBuffer
  return messageType.decode(new Uint8Array(data));
}

/** Encodes a JS object for transmission using the given protobuf message class */
export function encodeProtoPayload<T>(data: T, messageType: EncodableType<T>) {
  const encoded = messageType.encode(data).finish();
  const final = new Uint8Array(encoded.length);
  // ProtoBufJS uses a buffer pool, so the length of the encoded array will be
  // incorrect. We need to copy it into a new Uint8Array to fix it :-/
  for (let i = 0; i < encoded.length; i += 1) {
    final[i] = encoded[i];
  }
  return final;
}

/** Creates a an AdminEntityTransformer which converts a response to a
 * PaginatedEntityResponse by renaming one of the properties. `itemsKey`
 * specifies the name of the property to be renamed.
 * ex. `Admin.ExecutionsList` is of the shape { executions, token } and would be
 * converted to { entities, token }
 */
export function createPaginationTransformer<T, ResponseType extends { token?: string }>(
  itemsKey: keyof ResponseType,
): AdminEntityTransformer<ResponseType, PaginatedEntityResponse<T>> {
  return (response: ResponseType) => {
    return {
      token: response.token,
      entities: response[itemsKey] as unknown as T[],
    };
  };
}
