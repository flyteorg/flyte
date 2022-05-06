import { storiesOf } from '@storybook/react';
import { Scalar } from 'models/Common/types';
import { Card, CardContent } from '@material-ui/core';
import * as React from 'react';
import { Core } from 'flyteidl';
import { LiteralMapViewer } from '../LiteralMapViewer';
import { blobScalars } from './scalarValues';
import { extractSimpleTypes, extractSchemaTypes, extractBlobTypes } from './helpers/typeGenerators';

const stories = storiesOf('Literals/StructuredDataSet', module);

function generateStructuredDataset(label: string, literalType: Core.ILiteralType) {
  return {
    literals: {
      [label]: {
        scalar: {
          structuredDataset: {
            uri: 's3://random-val',
            metadata: {
              structuredDatasetType: {
                columns: [
                  {
                    name: 'age',
                    literalType,
                  },
                ],
                format: 'parquet',
              },
            },
          },
          value: 'structuredDataset',
        },
        value: 'scalar',
      },
    },
  };
}

function generateSimpleTypes() {
  const simpleTypes = extractSimpleTypes().map((simpleType) => {
    return Object.keys(simpleType).map((key) => {
      return generateStructuredDataset(`simple_${key}`, simpleType[key]);
    });
  });
  return simpleTypes.flat();
}

function generateSchemaTypes() {
  const schemaTypes = extractSchemaTypes();
  const retObj = schemaTypes
    .map((schemaType) => {
      return Object.keys(schemaType).map((key) => {
        const v = schemaType[key];
        return generateStructuredDataset(`schema${key}`, v);
      });
    })
    .flat();

  return retObj;
}

function generateBlobTypes() {
  return Object.keys(blobScalars).map((key) => {
    const value = blobScalars[key]?.['blob']?.metadata?.type;
    return generateStructuredDataset(key, {
      type: 'blob',
      blob: value,
    });
  });
}

function generateCollectionTypes() {
  const collectionTypes = [];

  const simpleTypes = extractSimpleTypes()
    .map((simpleType) => {
      return Object.keys(simpleType).map((key) => {
        return generateStructuredDataset(`simple_${key}`, {
          type: 'collectionType',
          collectionType: simpleType[key],
        });
      });
    })
    .flat();
  collectionTypes.push(...simpleTypes);

  // blobTypes
  const blobTypes = extractBlobTypes();
  collectionTypes.push(
    ...blobTypes
      .map((value) => {
        return Object.keys(value).map((key) => {
          return generateStructuredDataset(key, {
            type: 'collectionType',
            collectionType: value[key],
          });
        });
      })
      .flat(),
  );

  return collectionTypes;
}

function generateMapValueypes() {
  const mapValueTypes = [];

  const simpleTypes = extractSimpleTypes()
    .map((simpleType) => {
      return Object.keys(simpleType).map((key) => {
        return generateStructuredDataset(`simple_${key}`, {
          type: 'mapValueType',
          mapValueType: simpleType[key],
        });
      });
    })
    .flat();
  mapValueTypes.push(...simpleTypes);

  // blobTypes
  const blobTypes = extractBlobTypes();
  mapValueTypes.push(
    ...blobTypes
      .map((value) => {
        return Object.keys(value).map((key) => {
          return generateStructuredDataset(key, {
            type: 'mapValueType',
            mapValueType: value[key],
          });
        });
      })
      .flat(),
  );

  return mapValueTypes;
}

function generateEnumTypes() {
  const mapValueTypes = [
    generateStructuredDataset(`With_values`, {
      type: 'enumType',
      enumType: { values: ['1', '2', '3'] },
    }),
    generateStructuredDataset(`With_no_values`, {
      type: 'enumType',
      enumType: { values: [] },
    }),
    generateStructuredDataset(`With_null_values`, {
      type: 'enumType',
      enumType: { values: null },
    }),
  ];

  return mapValueTypes;
}

const renderScalars = (scalars: Dictionary<Scalar>) => {
  const items = Object.entries(scalars).map(([_label, value]) => {
    return (
      <Card style={{ margin: '16px 0', width: '430px' }}>
        <CardContent>
          <LiteralMapViewer map={value} />
        </CardContent>
      </Card>
    );
  });

  return <div style={{ display: 'flex', flexDirection: 'column' }}>{items}</div>;
};

stories.add('Simple types', () => <>{renderScalars(generateSimpleTypes())}</>);
stories.add('Blob Type', () => <>{renderScalars(generateBlobTypes())}</>);
stories.add('Schema types', () => <>{renderScalars(generateSchemaTypes())}</>);
stories.add('Collection Type', () => <>{renderScalars(generateCollectionTypes())}</>);
stories.add('mapValue Type', () => <>{renderScalars(generateMapValueypes())}</>);
stories.add('Enum Type', () => <>{renderScalars(generateEnumTypes())}</>);
