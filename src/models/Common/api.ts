import axios from 'axios';
import { Admin, Core } from 'flyteidl';
import {
    defaultPaginationConfig,
    getAdminEntity,
    getProfileUrl,
    PaginatedEntityResponse,
    RequestConfig
} from 'models/AdminEntity';

import { transformRequestError } from 'models/AdminEntity/transformRequestError';
import { defaultAxiosConfig, identifierPrefixes } from './constants';
import {
    IdentifierScope,
    NamedEntity,
    NamedEntityIdentifier,
    ResourceType,
    UserProfile
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
    requestConfig?: RequestConfig
) => {
    const prefix = identifierPrefixes[type];
    const path = scope ? makeIdentifierPath(prefix, scope) : prefix;

    return getAdminEntity<
        Admin.NamedEntityIdentifierList,
        PaginatedEntityResponse<NamedEntityIdentifier>
    >(
        {
            path,
            messageType: Admin.NamedEntityIdentifierList
        },
        { ...defaultPaginationConfig, ...requestConfig }
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
export const getNamedEntity = (
    input: GetNamedEntityInput,
    requestConfig?: RequestConfig
) => {
    return getAdminEntity<Admin.NamedEntity, NamedEntity>(
        {
            path: makeNamedEntityPath(input),
            messageType: Admin.NamedEntity
        },
        requestConfig
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
export const listNamedEntities = (
    input: ListNamedEntitiesInput,
    requestConfig?: RequestConfig
) => {
    const path = makeNamedEntityPath(input);

    return getAdminEntity<
        Admin.NamedEntityList,
        PaginatedEntityResponse<NamedEntity>
    >(
        {
            path,
            messageType: Admin.NamedEntityList
        },
        { ...defaultPaginationConfig, ...requestConfig }
    );
};

export const getUserProfile = async () => {
    const path = getProfileUrl();
    try {
        const { data } = await axios.get<UserProfile>(path, defaultAxiosConfig);
        return data;
    } catch (e) {
        const { message } = transformRequestError(e, path);
        console.error(`Failed to fetch user profile: ${message}`);
        return null;
    }
};
