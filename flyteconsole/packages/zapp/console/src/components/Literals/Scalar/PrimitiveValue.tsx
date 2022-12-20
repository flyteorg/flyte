import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { Primitive } from 'models/Common/types';
import * as React from 'react';

/** Stringifies a Primitive, handling any necessary formatting */
function primitiveToString(primitive: Primitive): string {
  switch (primitive.value) {
    case 'boolean':
      return primitive.boolean ? 'true' : 'false';
    case 'datetime':
      return formatDateUTC(timestampToDate(primitive.datetime!));
    case 'duration':
      return protobufDurationToHMS(primitive.duration!);
    default:
      return `${primitive[primitive.value]}`;
  }
}

export const PrimitiveValue: React.FC<{ primitive: Primitive }> = ({ primitive }) => {
  return <span>{primitiveToString(primitive)}</span>;
};
