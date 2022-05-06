import { Core } from 'flyteidl';
import { TestCaseList } from '../types';

const schemaColumnTypes: TestCaseList<Core.ISchema> = Object.keys(
  Core.SchemaType.SchemaColumn.SchemaColumnType,
)
  .map((key, index) => ({
    [`SCHEMA_WITH_${key}`]: {
      value: {
        uri: `s3/${index}`,
        type: {
          columns: [
            { name: 'name' + index, type: Core.SchemaType.SchemaColumn.SchemaColumnType[key] },
          ],
        },
      },
      expected: {
        result_var: { uri: `s3/${index}`, columns: [`name${index} (${key.toLocaleLowerCase()})`] },
      },
    },
  }))
  .reduce((acc, v) => {
    return {
      ...acc,
      ...v,
    };
  }, {});

const schemaTestCases: TestCaseList<Core.ISchema> = {
  SCHEMA_WITHOUT_TYPE: {
    value: {
      uri: 'test7',
      type: {
        columns: [{ name: 'test7' }],
      },
    },
    expected: {
      result_var: { uri: 'test7', columns: [`test7 (unknown)`] },
    },
  },
};

export default {
  ...schemaColumnTypes,
  ...schemaTestCases,
};
