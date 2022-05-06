import { Literal, LiteralCollection, LiteralMap, Scalar } from 'models/Common/types';
import * as React from 'react';
import { LiteralCollectionViewer } from './LiteralCollectionViewer';
import { DeprecatedLiteralMapViewer } from './DeprecatedLiteralMapViewer';
import { ScalarValue } from './Scalar/ScalarValue';
import { useLiteralStyles } from './styles';
import { UnsupportedType } from './UnsupportedType';
import { ValueLabel } from './ValueLabel';

/** Renders a Literal as a formatted label/value */
export const LiteralValue: React.FC<{
  label: React.ReactNode;
  literal: Literal;
}> = ({ label, literal }) => {
  const literalStyles = useLiteralStyles();
  switch (literal.value) {
    case 'scalar':
      return <ScalarValue label={label} scalar={literal.scalar as Scalar} />;
    case 'collection':
      return (
        <>
          <ValueLabel label={label} />
          <LiteralCollectionViewer collection={literal.collection as LiteralCollection} />
        </>
      );
    case 'map':
      return (
        <>
          <ValueLabel label={label} />
          <DeprecatedLiteralMapViewer
            className={literalStyles.nestedContainer}
            map={literal.map as LiteralMap}
            showBrackets={true}
          />
        </>
      );
    default:
      return <UnsupportedType label={label} />;
  }
};
