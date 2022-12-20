import { Protobuf } from 'flyteidl';
import { TestCaseList } from '../types';
import { getIValue } from './literalHelpers';

const nullValueTestcases: TestCaseList<Protobuf.IStruct> = {
  WITH_NULL_VAL: {
    value: {
      fields: {
        test_field_name1: getIValue('nullValue', Protobuf.NullValue.NULL_VALUE),
      },
    },
    expected: {
      result_var: { test_field_name1: '(empty)' },
    },
  },
};
const numberValueTestCases: TestCaseList<Protobuf.IStruct> = {
  WITH_NUMBER_VAL: {
    value: {
      fields: {
        test_field_name2: getIValue('numberValue', 1),
      },
    },
    expected: {
      result_var: { test_field_name2: 1 },
    },
  },
  WITH_NUMBER_VAL_NULL: {
    value: {
      fields: {
        test_field_name3: getIValue('numberValue', null),
      },
    },
    expected: {
      result_var: { test_field_name3: null },
    },
  },
};
const stringValueTestCases: TestCaseList<Protobuf.IStruct> = {
  WITH_STRING_VAL: {
    value: {
      fields: {
        test_field_name4: getIValue('stringValue', 'test val'),
      },
    },
    expected: {
      result_var: { test_field_name4: 'test val' },
    },
  },
  WITH_STRING_VAL_NULL: {
    value: {
      fields: {
        test_field_name: getIValue('stringValue', null),
      },
    },
    expected: {
      result_var: { test_field_name: null },
    },
  },
};

const boolValueTestCases: TestCaseList<Protobuf.IStruct> = {
  WITH_BOOL_VAL: {
    value: {
      fields: {
        test_field_name: getIValue('boolValue', true),
      },
    },
    expected: {
      result_var: { test_field_name: true },
    },
  },
  WITH_BOOL_VAL_FALSE: {
    value: {
      fields: {
        test_field_name: getIValue('boolValue', false),
      },
    },
    expected: {
      result_var: { test_field_name: false },
    },
  },
  WITH_BOOL_VAL_NULL: {
    value: {
      fields: {
        test_field_name: getIValue('boolValue', null),
      },
    },
    expected: {
      result_var: { test_field_name: null },
    },
  },
};

const structValueTestCases: TestCaseList<Protobuf.IStruct> = {
  WITH_STRUCT_VALUE: {
    value: {
      fields: {
        test_struct_name: getIValue('structValue', {
          fields: {
            struct_string_val_copy: stringValueTestCases.WITH_STRING_VAL.value?.fields
              ?.test_field_name4 as Protobuf.IValue,
            struct_bool_val_copy: boolValueTestCases.WITH_BOOL_VAL.value?.fields
              ?.test_field_name as Protobuf.IValue,
          },
        }),
      },
    },
    expected: {
      result_var: {
        test_struct_name: { struct_string_val_copy: 'test val', struct_bool_val_copy: true },
      },
    },
  },
};

const listValueTestCases: TestCaseList<Protobuf.IStruct> = {
  WITH_LIST_VALUE: {
    value: {
      fields: {
        test_list_name: getIValue('listValue', {
          values: [
            structValueTestCases.WITH_STRUCT_VALUE.value?.fields
              ?.test_struct_name as Protobuf.IValue,
          ],
        }),
      },
    },
    expected: {
      result_var: {
        test_list_name: [{ struct_bool_val_copy: true, struct_string_val_copy: 'test val' }],
      },
    },
  },
};
export default {
  ...nullValueTestcases,
  ...numberValueTestCases,
  ...stringValueTestCases,
  ...boolValueTestCases,
  ...structValueTestCases,
  ...listValueTestCases,
};
