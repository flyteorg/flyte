import axios, { AxiosRequestConfig, Method } from 'axios';
import { generateAdminApiQuery } from './AdminApiQuery';
import { transformRequestError } from './transformRequestError';
import { AdminEntityTransformer, DecodableType, EncodableType, RequestConfig } from './types';
import { adminApiUrl, decodeProtoResponse, encodeProtoPayload, logProtoResponse } from './utils';

/** Base work function used by the HTTP verb methods below. It does not handle
 * encoding/decoding of protobuf.
 */
async function request(
  /** HTTP verb to use */
  method: Method,
  /** API endpoint to use, should not include protocol/host/prefix */
  endpoint: string,
  /** Admin API options to use for the request */
  config: RequestConfig = {},
) {
  const options: AxiosRequestConfig = {
    method,
    data: config.data,
  };

  options.params = { ...config.params, ...generateAdminApiQuery(config) };

  /* For protobuf responses, we need special accept/content headers and
    responseType */
  options.headers = { Accept: 'application/octet-stream' };
  options.responseType = 'arraybuffer';
  if (config.data) {
    options.headers['Content-Type'] = 'application/octet-stream';
  }

  const finalOptions = {
    ...options,
    url: adminApiUrl(endpoint),
    withCredentials: true,
  };

  try {
    const { data } = await axios.request(finalOptions);
    return data;
  } catch (e) {
    throw transformRequestError(e, endpoint);
  }
}

/** A generic getter function for fetching protobuf data from a given URL.
 * @param config - A standard `AxiosRequestConfig`. `url` is required.
 * @param messageType - A protobuf message class to use for decoding
 */
export async function getProtobufObject<ResponseType>(
  config: AxiosRequestConfig & { url: string },
  messageType: DecodableType<ResponseType>,
) {
  const { headers = {}, ...restOptions } = config;
  headers.Accept = 'application/octet-stream';
  const options: AxiosRequestConfig = {
    ...restOptions,
    headers,
    method: 'get',
    responseType: 'arraybuffer',
    withCredentials: true,
  };

  const { data } = await axios.request(options);
  const decoded = decodeProtoResponse(data, messageType);
  logProtoResponse(config.url, decoded);
  return decoded;
}

export interface GetEntityParams<T, TransformedType> {
  path: string;
  messageType: DecodableType<T>;
  transform?: AdminEntityTransformer<T, TransformedType>;
}

function identityTransformer(msg: any) {
  return msg;
}

/** GETs an entity by path and decodes/transforms it using provided functions */
export async function getAdminEntity<ResponseType, TransformedType>(
  {
    path,
    messageType,
    transform = identityTransformer,
  }: GetEntityParams<ResponseType, TransformedType>,
  config?: RequestConfig,
): Promise<TransformedType> {
  const data: ArrayBuffer = await request('get', path, config);
  const decoded = decodeProtoResponse(data, messageType);
  logProtoResponse(path, decoded);
  return transform(decoded) as TransformedType;
}

export interface PostEntityParams<RequestType, ResponseType, TransformedType> {
  data: RequestType;
  path: string;
  method?: Method;
  requestMessageType: EncodableType<RequestType>;
  responseMessageType: DecodableType<ResponseType>;
  transform?: AdminEntityTransformer<ResponseType, TransformedType>;
}

/** POSTs an entity, encoded as protobuf, by path and decodes/transforms it
 * using provided request and response message types.
 */
export async function postAdminEntity<RequestType, ResponseType, TransformedType = ResponseType>(
  {
    path,
    data,
    method = 'post',
    requestMessageType,
    responseMessageType,
    transform = identityTransformer,
  }: PostEntityParams<RequestType, ResponseType, TransformedType>,
  config?: RequestConfig,
): Promise<TransformedType> {
  const body = encodeProtoPayload(data, requestMessageType);
  const finalConfig = { ...config, data: body };
  const responseData: ArrayBuffer = await request(method, path, finalConfig);
  const decoded = decodeProtoResponse(responseData, responseMessageType);
  logProtoResponse(path, decoded);
  return transform(decoded) as TransformedType;
}
