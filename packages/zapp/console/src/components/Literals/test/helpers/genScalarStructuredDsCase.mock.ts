import { Core } from 'flyteidl';
import { TestCaseList } from '../types';
import { processSimpleType, columnTypeToString } from '../../helpers';

import { simple } from './mock_simpleTypes';

const generateColumnEntry = (columnName: string, literalType) => {
  return {
    name: columnName,
    literalType,
  };
};

const generateStructuredDataset = (columnName: string, uri: string, literalType) => {
  return {
    uri,
    metadata: {
      structuredDatasetType: {
        format: 'parquet',
        columns: [generateColumnEntry(columnName, literalType)],
      },
    },
  };
};

const sasWithMapValueTypeColumns: TestCaseList<Core.IStructuredDataset> = Object.keys(simple)
  .map((simpleTypeKey, index) => {
    const simpleType = simple[simpleTypeKey];

    const literalType = {
      mapValueType: {
        ...simpleType,
      },
      type: 'mapValueType',
    };

    const name = `column_name_${index}`;
    const columns = [];
    columns[name] = `map value of ${processSimpleType(simpleType[simpleType.type])}`;

    return {
      [`STRUCT_SIMPLE_${simpleTypeKey}`]: {
        value: generateStructuredDataset(name, index.toString(), literalType),
        expected: {
          result_var: { columns, format: 'parquet', uri: index.toString() },
        },
      },
    };
  })
  .reduce((acc, v) => ({ ...acc, ...v }), {});

const sasWithCollectionTypeColumns: TestCaseList<Core.IStructuredDataset> = Object.keys(simple)
  .map((simpleTypeKey, index) => {
    const simpleType = simple[simpleTypeKey];

    const literalType = {
      collectionType: {
        ...simpleType,
      },
      type: 'collectionType',
    };

    const name = `column_name_${index}`;
    const columns = [];
    columns[name] = `collection of ${processSimpleType(simpleType[simpleType.type])}`;

    return {
      [`STRUCT_SIMPLE_${simpleTypeKey}`]: {
        value: generateStructuredDataset(name, index.toString(), literalType),
        expected: {
          result_var: { columns, format: 'parquet', uri: index.toString() },
        },
      },
    };
  })
  .slice(0, 1)
  .reduce((acc, v) => ({ ...acc, ...v }), {});

const sdsWithSimpleTypeColumns: TestCaseList<Core.IStructuredDataset> = Object.keys(simple)
  .map((simpleTypeKey, index) => {
    const literalType = simple[simpleTypeKey];

    const name = `column_name_${index}`;
    const columns = [];
    columns[name] = processSimpleType(literalType[literalType.type]);

    return {
      [`STRUCT_SIMPLE_${simpleTypeKey}`]: {
        value: generateStructuredDataset(name, index.toString(), literalType),
        expected: {
          result_var: { columns, format: 'parquet', uri: index.toString() },
        },
      },
    };
  })
  .reduce((acc, v) => ({ ...acc, ...v }), {});

const sasWithSchemaColumns: TestCaseList<Core.IStructuredDataset> = Object.keys(
  Core.SchemaType.SchemaColumn.SchemaColumnType,
)
  .map((simpleTypeKey, index) => {
    const literalType = {
      schema: {
        columns: [{ type: Core.SchemaType.SchemaColumn.SchemaColumnType[simpleTypeKey] }],
      },
      type: 'schema',
    };

    const name = `schema_column_name_${index}`;
    const columns = [];
    columns[name] = `schema (${columnTypeToString(
      Core.SchemaType.SchemaColumn.SchemaColumnType[simpleTypeKey],
    )})`;

    return {
      [`STRUCT_SCHEMA_WITH_COLUMNS_${simpleTypeKey}`]: {
        value: generateStructuredDataset(name, index.toString(), literalType),
        expected: {
          result_var: { columns, format: 'parquet', uri: index.toString() },
        },
      },
    };
  })
  .reduce((acc, v) => ({ ...acc, ...v }), {});

export default {
  ...sdsWithSimpleTypeColumns,
  ...sasWithSchemaColumns,
  ...sasWithCollectionTypeColumns,
  ...sasWithMapValueTypeColumns,
};
