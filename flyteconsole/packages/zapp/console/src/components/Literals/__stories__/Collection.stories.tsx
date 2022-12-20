import { storiesOf } from '@storybook/react';
import { LiteralCollection } from 'models/Common/types';
import * as React from 'react';
import { Card, CardContent } from '@material-ui/core';
import { LiteralMapViewer } from '../LiteralMapViewer';
import { DeprecatedLiteralMapViewer } from '../DeprecatedLiteralMapViewer';
import {
  binaryLiterals,
  blobLiterals,
  errorLiterals,
  noneTypeLiteral,
  primitiveLiterals,
  schemaLiterals,
} from './literalValues';

const stories = storiesOf('Literals/Collection', module);

function renderCollection(label: string, collection: LiteralCollection) {
  const map = { literals: { [label]: { collection, value: 'collection' } } };
  return (
    <>
      <div style={{ display: 'flex' }}>
        <div style={{ marginRight: '16px' }}>
          OLD
          <Card>
            <CardContent>
              <DeprecatedLiteralMapViewer map={map} />
            </CardContent>
          </Card>
        </div>
        <div>
          NEW
          <Card>
            <CardContent>
              <LiteralMapViewer map={map} />
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
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
