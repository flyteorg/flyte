import { Schema, SchemaColumnType } from 'models/Common/types';

const schemaWithNoColumns: Schema = {
  uri: 's3://path/to/my/schema_with_no_columns',
  type: {
    columns: [],
  },
};

const schemaWithColumns: Schema = {
  uri: 's3://path/to/my/schema_with_no_columns',
  type: {
    columns: [
      { name: 'booleanColumn', type: SchemaColumnType.BOOLEAN },
      { name: 'datetimeColumn', type: SchemaColumnType.DATETIME },
      { name: 'durationColumn', type: SchemaColumnType.DURATION },
      { name: 'floatColumn', type: SchemaColumnType.FLOAT },
      { name: 'stringColumn', type: SchemaColumnType.STRING },
      { name: 'integerColumn', type: SchemaColumnType.INTEGER },
    ],
  },
};

export const schemaValues = { schemaWithColumns, schemaWithNoColumns };
