import { Core } from 'flyteidl';
import {
  blobScalars,
  schemaScalars,
} from '../scalarValues';

// SIMPLE
type GeneratedSimpleType = {
  [key in Core.SimpleType]?: Core.LiteralType;
};
export function extractSimpleTypes() {
  const simpleTypes: GeneratedSimpleType[] = Object.keys(Core.SimpleType).map((key) => ({
    [key]: {
      type: 'simple',
      simple: Core.SimpleType[key],
    },
  }));
  return simpleTypes;
}

// SCHEMA
export function extractSchemaTypes() {
  return Object.keys(schemaScalars).map((key) => {
    const value = schemaScalars[key];
    return {
      [key]: {
        type: 'schema',
        schema: value?.schema?.type,
      },
    };
  });
}

// COLLECTION TYPE
export function extractBlobTypes() {
  return Object.keys(blobScalars).map((key) => {
    const value = blobScalars[key]?.['blob']?.metadata?.type;
    return {
      [key]: {
        type: 'blob',
        blob: value,
      },
    };
  });
}
