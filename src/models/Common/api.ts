import axios from 'axios';
import { env } from 'common/env';
import { log } from 'common/log';
import { createCorsProxyURL } from 'common/utils';
import { Admin, Core } from 'flyteidl';
import { getAdminEntity, getProtobufObject } from 'models/AdminEntity/AdminEntity';
import { defaultPaginationConfig } from 'models/AdminEntity/constants';
import { transformRequestError } from 'models/AdminEntity/transformRequestError';
import { PaginatedEntityResponse, RequestConfig } from 'models/AdminEntity/types';
import { adminApiUrl, getProfileUrl } from 'models/AdminEntity/utils';
import { defaultAxiosConfig, defaultSystemStatus, identifierPrefixes } from './constants';
import {
  GetVersionResponse,
  IdentifierScope,
  LiteralMap,
  NamedEntity,
  NamedEntityIdentifier,
  ResourceType,
  SystemStatus,
  UserProfile,
} from './types';
import { makeIdentifierPath, makeNamedEntityPath } from './utils';

interface ListIdentifiersConfig {
  type: ResourceType;
  scope?: IdentifierScope;
}

/** Fetches a list of NamedEntityIdentifiers from the Admin API
 * @param config Controls the query for the list request
 * @param config.type Specifies which type of resource IDs to query
 * @param config.scope A partial Identifier used to scope the results. Values
 * are applied in hierarchical order (project, domain, name) and it is invalid
 * to pass a lower-hierarchy scope (ex. name) without also specifying the
 * levels above it (project and domain)
 * @param requestConfig A standard `RequestConfig` object
 */
export const listIdentifiers = (
  { type, scope }: ListIdentifiersConfig,
  requestConfig?: RequestConfig,
) => {
  const prefix = identifierPrefixes[type];
  const path = scope ? makeIdentifierPath(prefix, scope) : prefix;

  return getAdminEntity<
    Admin.NamedEntityIdentifierList,
    PaginatedEntityResponse<NamedEntityIdentifier>
  >(
    {
      path,
      messageType: Admin.NamedEntityIdentifierList,
    },
    { ...defaultPaginationConfig, ...requestConfig },
  );
};

export interface GetNamedEntityInput {
  resourceType: Core.ResourceType;
  project: string;
  domain: string;
  name: string;
}

/** Fetches a NamedEntity from the Admin API
 * @param input An object specifying the resource type, project, domain, and
 * name of the entity to fetch. All fields are _required_
 * @param requestConfig A standard `RequestConfig` object
 */
export const getNamedEntity = (input: GetNamedEntityInput, requestConfig?: RequestConfig) => {
  return getAdminEntity<Admin.NamedEntity, NamedEntity>(
    {
      path: makeNamedEntityPath(input),
      messageType: Admin.NamedEntity,
    },
    requestConfig,
  );
};

export interface ListNamedEntitiesInput {
  resourceType: Core.ResourceType;
  project: string;
  domain: string;
}

/** Fetches a list of NamedEntity objects sharing a common project/domain
 * @param input An object specifying the resource type, project, and domain.
 * All fields are _required_
 * @param requestConfig A standard `RequestConfig` object
 */
export const listNamedEntities = (input: ListNamedEntitiesInput, requestConfig?: RequestConfig) => {
  const path = makeNamedEntityPath(input);

  return getAdminEntity<Admin.NamedEntityList, PaginatedEntityResponse<NamedEntity>>(
    {
      path,
      messageType: Admin.NamedEntityList,
    },
    { ...defaultPaginationConfig, ...requestConfig },
  );
};

/** Fetches the current user profile. NOTE: This will *not* fail in cases
 * where the user is not logged in or the session is expired. Admin does not
 * distinguish between these cases, so the profile will be `null` in both cases.
 * A value of `null` indicates that a redirect to the login endpoint is needed.
 */
export const getUserProfile = async () => {
  const path = getProfileUrl();
  try {
    const { data } = await axios.get<UserProfile>(path, defaultAxiosConfig);
    return data;
  } catch (e) {
    const { message } = transformRequestError(e, path);
    log.error(`Failed to fetch user profile: ${message}`);
    return null;
  }
};

/** Fetches the current admin version.
 */
export const getVersion = async () => {
  const path = adminApiUrl('/version');
  try {
    const { data } = await axios.get<GetVersionResponse>(path, defaultAxiosConfig);
    return data;
  } catch (e) {
    const { message } = transformRequestError(e, path);
    log.error(`Failed to fetch version: ${message}`);
    return null;
  }
};

/** If env.STATUS_URL is set, will issue a fetch to retrieve the current system
 * status. If not, will resolve immediately with a default value indicating
 * normal system status.
 */
export const getSystemStatus = async () => {
  if (!env.STATUS_URL) {
    return defaultSystemStatus;
  }
  const path = env.STATUS_URL;

  try {
    const { data } = await axios.get<SystemStatus>(path, defaultAxiosConfig);
    return data;
  } catch (e) {
    const { message } = transformRequestError(e, path);
    throw new Error(`Failed to fetch system status: ${message}`);
  }
};

/** Given a url to a `LiteralMap` stored at a remote location, will fetch and
 * decode the resulting object.
 */
export const getRemoteLiteralMap = async (url: string) =>
  getProtobufObject<LiteralMap>({ url: createCorsProxyURL(url) }, LiteralMap);
