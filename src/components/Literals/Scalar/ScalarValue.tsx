import {
  Binary,
  Blob,
  Error,
  Primitive,
  ProtobufStruct,
  Scalar,
  Schema,
} from 'models/Common/types';
import * as React from 'react';
import { PrintValue } from '../PrintValue';
import { UnsupportedType } from '../UnsupportedType';
import { ValueLabel } from '../ValueLabel';
import { BinaryValue } from './BinaryValue';
import { BlobValue } from './BlobValue';
import { ErrorValue } from './ErrorValue';
import { NoneTypeValue } from './NoneTypeValue';
import { PrimitiveValue } from './PrimitiveValue';
import { ProtobufStructValue } from './ProtobufStructValue';
import { SchemaValue } from './SchemaValue';

/** Renders a `Scalar` using appropriate sub-components for each possible type */
export const ScalarValue: React.FC<{
  label: React.ReactNode;
  scalar: Scalar;
}> = ({ label, scalar }) => {
  switch (scalar.value) {
    case 'primitive':
      return (
        <PrintValue
          label={label}
          value={<PrimitiveValue primitive={scalar.primitive as Primitive} />}
        />
      );
    case 'blob':
      return (
        <>
          <ValueLabel label={label} />
          <BlobValue blob={scalar.blob as Blob} />
        </>
      );
    case 'schema':
      return (
        <>
          <ValueLabel label={label} />
          <SchemaValue schema={scalar.schema as Schema} />
        </>
      );
    case 'error':
      return (
        <>
          <ValueLabel label={label} />
          <ErrorValue error={scalar.error as Error} />
        </>
      );
    case 'binary':
      return (
        <>
          <ValueLabel label={label} />
          <BinaryValue binary={scalar.binary as Binary} />
        </>
      );
    case 'generic':
      return (
        <>
          <ValueLabel label={label} />
          <ProtobufStructValue struct={scalar.generic as ProtobufStruct} />
        </>
      );
    case 'noneType':
      return <PrintValue label={label} value={<NoneTypeValue />} />;
    default:
      return <UnsupportedType label={label} />;
  }
};
