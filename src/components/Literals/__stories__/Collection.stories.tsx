import { storiesOf } from '@storybook/react';
import { LiteralCollection } from 'models/Common/types';
import * as React from 'react';
import { LiteralValue } from '../LiteralValue';
import { CardDecorator } from './CardDecorator';
import {
  binaryLiterals,
  blobLiterals,
  errorLiterals,
  noneTypeLiteral,
  primitiveLiterals,
  schemaLiterals,
} from './literalValues';

const stories = storiesOf('Literals/Collection', module);
stories.addDecorator(CardDecorator);

function renderCollection(label: string, collection: LiteralCollection) {
  return <LiteralValue label={label} literal={{ collection, value: 'collection', hash: '' }} />;
}

stories.add('Binary', () =>
  renderCollection('binary_collection', {
    literals: Object.values(binaryLiterals),
  }),
);
stories.add('Blob', () =>
  renderCollection('blob_collection', {
    literals: Object.values(blobLiterals),
  }),
);
stories.add('Error', () =>
  renderCollection('error_collection', {
    literals: Object.values(errorLiterals),
  }),
);
stories.add('NoneType', () =>
  renderCollection('noneType_collection', {
    literals: [noneTypeLiteral],
  }),
);
stories.add('Primitive', () =>
  renderCollection('primitive_collection', {
    literals: Object.values(primitiveLiterals),
  }),
);
stories.add('Schema', () =>
  renderCollection('schema_collection', {
    literals: Object.values(schemaLiterals),
  }),
);
stories.add('Mixed', () =>
  renderCollection('mixed_collection', {
    literals: [
      ...Object.values(binaryLiterals),
      ...Object.values(blobLiterals),
      ...Object.values(errorLiterals),
      noneTypeLiteral,
      ...Object.values(primitiveLiterals),
      ...Object.values(schemaLiterals),
    ],
  }),
);
stories.add('Empty', () => renderCollection('empty_collection', { literals: [] }));
