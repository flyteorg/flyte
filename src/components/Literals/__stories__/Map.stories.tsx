import { storiesOf } from '@storybook/react';
import { LiteralMap } from 'models/Common/types';
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

const stories = storiesOf('Literals/Map', module);
stories.addDecorator(CardDecorator);

function renderMap(label: string, map: LiteralMap) {
  return <LiteralValue label={label} literal={{ map, value: 'map' }} />;
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
