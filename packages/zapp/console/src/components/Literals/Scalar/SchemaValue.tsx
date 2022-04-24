import { Schema, SchemaColumn, SchemaColumnType } from 'models/Common/types';
import * as React from 'react';
import { PrintList } from '../PrintList';
import { PrintValue } from '../PrintValue';
import { useLiteralStyles } from '../styles';
import { ValueLabel } from '../ValueLabel';

function columnTypeToString(type: SchemaColumnType) {
  switch (type) {
    case SchemaColumnType.BOOLEAN:
      return 'boolean';
    case SchemaColumnType.DATETIME:
      return 'datetime';
    case SchemaColumnType.DURATION:
      return 'duration';
    case SchemaColumnType.FLOAT:
      return 'float';
    case SchemaColumnType.INTEGER:
      return 'integer';
    case SchemaColumnType.STRING:
      return 'string';
    default:
      return 'unknown';
  }
}

function getColumnKey(value: any) {
  return (value as SchemaColumn).name;
}

function renderSchemaColumn(value: any, index: number) {
  const { name, type } = value as SchemaColumn;
  return <PrintValue label={index} value={`${name} (${columnTypeToString(type)})`} />;
}

/** Renders a `Schema` definition as an object with a uri and optional list
 * of columns
 */
export const SchemaValue: React.FC<{ schema: Schema }> = ({ schema }) => {
  const literalStyles = useLiteralStyles();
  let columns;
  if (schema.type.columns.length > 0) {
    columns = (
      <>
        <ValueLabel label="columns" />
        <PrintList
          values={schema.type.columns}
          getKeyForItem={getColumnKey}
          renderValue={renderSchemaColumn}
        />
      </>
    );
  }

  return (
    <div className={literalStyles.nestedContainer}>
      <PrintValue label="uri" value={schema.uri} />
      {columns}
    </div>
  );
};
