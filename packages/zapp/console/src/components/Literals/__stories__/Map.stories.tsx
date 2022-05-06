import { storiesOf } from '@storybook/react';
import { LiteralMap } from 'models/Common/types';
import * as React from 'react';
import { Card, CardContent } from '@material-ui/core';
import {
  binaryLiterals,
  blobLiterals,
  errorLiterals,
  noneTypeLiteral,
  primitiveLiterals,
  schemaLiterals,
} from './literalValues';
import { LiteralMapViewer } from '../LiteralMapViewer';
import { DeprecatedLiteralMapViewer } from '../DeprecatedLiteralMapViewer';

const stories = storiesOf('Literals/Map', module);

function renderMap(label: string, map: LiteralMap) {
  const fullMap = { literals: { [label]: { map, value: 'map' } } };
  return (
    <>
      <div style={{ display: 'flex' }}>
        <div style={{ marginRight: '16px' }}>
          OLD
          <Card>
            <CardContent>
              <DeprecatedLiteralMapViewer map={fullMap} />
            </CardContent>
          </Card>
        </div>
        <div>
          NEW
          <Card>
            <CardContent>
              <LiteralMapViewer map={fullMap} />
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

stories.add('Binary', () =>
  renderMap('binary_map', {
    literals: binaryLiterals,
  }),
);
stories.add('Blob', () =>
  renderMap('blob_map', {
    literals: blobLiterals,
  }),
);
stories.add('Error', () =>
  renderMap('error_map', {
    literals: errorLiterals,
  }),
);
stories.add('NoneType', () =>
  renderMap('noneType_map', {
    literals: { empty_value: noneTypeLiteral },
  }),
);
stories.add('Primitive', () =>
  renderMap('primitive_map', {
    literals: primitiveLiterals,
  }),
);
stories.add('Schema', () =>
  renderMap('schema_map', {
    literals: schemaLiterals,
  }),
);
stories.add('Mixed', () =>
  renderMap('mixed_map', {
    literals: {
      ...binaryLiterals,
      ...blobLiterals,
      ...errorLiterals,
      noneTypeLiteral,
      ...primitiveLiterals,
      ...schemaLiterals,
    },
  }),
);
stories.add('Empty', () => renderMap('empty_map', { literals: {} }));
