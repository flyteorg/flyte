import { Admin } from 'flyteidl';
import {
    defaultPaginationConfig,
    getAdminEntity,
    PaginatedEntityResponse,
    RequestConfig
} from 'models/AdminEntity';

import { identifierPrefixes } from './constants';
import { IdentifierScope, NamedEntityIdentifier, ResourceType } from './types';
import { makeIdentifierPath } from './utils';

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
