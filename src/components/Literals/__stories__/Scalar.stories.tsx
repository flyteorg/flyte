import { storiesOf } from '@storybook/react';
import { Scalar } from 'models/Common/types';
import * as React from 'react';
import { ScalarValue } from '../Scalar/ScalarValue';
import { CardDecorator } from './CardDecorator';
import {
  binaryScalars,
  blobScalars,
  errorScalars,
  noneTypeScalar,
  primitiveScalars,
  schemaScalars,
} from './scalarValues';

const stories = storiesOf('Literals/Scalar', module);
stories.addDecorator(CardDecorator);

const renderScalars = (scalars: Dictionary<Scalar>) =>
  Object.entries(scalars).map(([label, value]) => {
    return <ScalarValue key={label} label={label} scalar={value as Scalar} />;
  });

stories.add('Binary', () => <>{renderScalars(binaryScalars)}</>);
stories.add('Blob', () => <>{renderScalars(blobScalars)}</>);
stories.add('Error', () => <>{renderScalars(errorScalars)}</>);
stories.add('NoneType', () => (
  <>
    <ScalarValue label="empty_value" scalar={noneTypeScalar} />
  </>
));
stories.add('Primitive', () => <>{renderScalars(primitiveScalars)}</>);
stories.add('Schema', () => <>{renderScalars(schemaScalars)}</>);
