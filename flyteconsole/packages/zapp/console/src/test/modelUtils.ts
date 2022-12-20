import { Core } from 'flyteidl';
import {
  Identifier,
  NamedEntity,
  NamedEntityIdentifier,
  NamedEntityMetadata,
  ResourceType,
} from 'models/Common/types';
import { NamedEntityState } from 'models/enums';

const defaultMetadata = {
  description: '',
  state: NamedEntityState.NAMED_ENTITY_ACTIVE,
};

export function createNamedEntity(
  resourceType: ResourceType,
  id: NamedEntityIdentifier,
  metadataOverrides?: Partial<NamedEntityMetadata>,
): NamedEntity {
  return {
    id,
    resourceType,
    metadata: { ...defaultMetadata, ...metadataOverrides },
  };
}

export function makeIdentifier(id?: Partial<Identifier>): Identifier {
  return {
    resourceType: Core.ResourceType.UNSPECIFIED,
    project: 'project',
    domain: 'domain',
    name: 'name',
    version: 'version',
    ...id,
  };
}

export function createWorkflowName(
  id: NamedEntityIdentifier,
  metadata?: Partial<NamedEntityMetadata>,
) {
  return createNamedEntity(ResourceType.WORKFLOW, id, metadata);
}
