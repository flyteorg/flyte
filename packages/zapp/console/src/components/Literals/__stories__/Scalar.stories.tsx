import { storiesOf } from '@storybook/react';
import { Scalar } from 'models/Common/types';
import * as React from 'react';
import { Card, CardContent } from '@material-ui/core';
import { LiteralMapViewer } from '../LiteralMapViewer';
import {
  binaryScalars,
  blobScalars,
  errorScalars,
  noneTypeScalar,
  primitiveScalars,
  schemaScalars,
} from './scalarValues';
import { DeprecatedLiteralMapViewer } from '../DeprecatedLiteralMapViewer';

const stories = storiesOf('Literals/Scalar', module);

var renderScalars = (scalars: Dictionary<Scalar>) => {
  const literals = {};

  Object.entries(scalars).map(([label, value]) => {
    literals[label] = { scalar: value, value: 'scalar' };
  });

  const map = { literals };
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
};

stories.add('Binary', () => <>{renderScalars(binaryScalars)}</>);
stories.add('Blob', () => <>{renderScalars(blobScalars)}</>);
stories.add('Error', () => <>{renderScalars(errorScalars)}</>);
stories.add('NoneType', () => <>{renderScalars([noneTypeScalar])}</>);
stories.add('Primitive', () => <>{renderScalars(primitiveScalars)}</>);
stories.add('Schema', () => <>{renderScalars(schemaScalars)}</>);
