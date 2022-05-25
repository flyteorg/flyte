import axios, {
  AxiosRequestConfig,
  AxiosRequestTransformer,
  AxiosResponseTransformer,
} from 'axios';
import * as snakecaseKeys from 'snakecase-keys';
import * as camelcaseKeys from 'camelcase-keys';
import { AdminEndpoint, adminApiPrefix, RawEndpoint } from './constants';
import { isObject } from './nodeChecks';
import { transformRequestError } from './errors';

/** Ensures that a string is slash-prefixed */
function ensureSlashPrefixed(path: string) {
  return path.startsWith('/') ? path : `/${path}`;
}

/** Creates a URL to the same host with a given path */
function createLocalURL(path: string) {
  return `${window.location.origin}${ensureSlashPrefixed(path)}`;
}

/** Updates Enpoint url depending on admin domain */
export function getEndpointUrl(endpoint: RawEndpoint | string, adminUrl?: string) {
  if (adminUrl) {
    return `${adminUrl}${endpoint}`;
  }

  return createLocalURL(endpoint);
}

/** Adds admin api prefix to domain Url */
export function getAdminApiUrl(endpoint: AdminEndpoint | string, adminUrl?: string) {
  const finalUrl = `${adminApiPrefix}${ensureSlashPrefixed(endpoint)}`;

  if (adminUrl) {
    return `${adminUrl}${finalUrl}`;
  }

  return createLocalURL(finalUrl);
}

/** Config object that can be used for requests that are not sent to
 * the Admin entity API (`/api/v1/...`), such as the `/me` endpoint. This config
 * ensures that requests/responses are correctly converted and that cookies are
 * included.
 */
export const defaultAxiosConfig: AxiosRequestConfig = {
  transformRequest: [
    (data: any) => (isObject(data) ? snakecaseKeys(data) : data),
    ...(axios.defaults.transformRequest as AxiosRequestTransformer[]),
  ],
  transformResponse: [
    ...(axios.defaults.transformResponse as AxiosResponseTransformer[] as any),
    camelcaseKeys,
  ],
  withCredentials: true,
};

/**
 * @deprecated Please use `axios-hooks` instead, it will allow you to get full call status.
 * example usage https://www.npmjs.com/package/axios-hooks:
 * const [{ data: profile, loading, errot }] = useAxios({url: path, method: 'GET', ...defaultAxiosConfig});
 */
export const getAxiosApiCall = async <T>(path: string): Promise<T | null> => {
  try {
    const { data } = await axios.get<T>(path, defaultAxiosConfig);
    return data;
  } catch (e) {
    const { message } = transformRequestError(e, path);
    // eslint-disable-next-line no-console
    console.error(`Failed to fetch data: ${message}`);
    return null;
  }
};
