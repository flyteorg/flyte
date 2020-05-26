import { Admin } from 'flyteidl';
import {
    NamedEntity,
    NamedEntityIdentifier,
    NamedEntityMetadata,
    ResourceType
} from 'models';

const defaultMetadata = {
    description: '',
    state: Admin.NamedEntityState.NAMED_ENTITY_ACTIVE
};

function createNamedEntity(
    resourceType: ResourceType,
    id: NamedEntityIdentifier,
    metadataOverrides?: Partial<NamedEntityMetadata>
): NamedEntity {
    return {
        id,
        resourceType,
        metadata: { ...defaultMetadata, ...metadataOverrides }
    };
}

export function createWorkflowName(
    id: NamedEntityIdentifier,
    metadata?: Partial<NamedEntityMetadata>
) {
    return createNamedEntity(ResourceType.WORKFLOW, id, metadata);
}
