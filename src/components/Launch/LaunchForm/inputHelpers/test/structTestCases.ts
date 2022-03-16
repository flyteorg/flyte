import { stringifyValue } from 'common/utils';
import { Core, Protobuf } from 'flyteidl';

export function structLiteral(generic: Protobuf.IStruct): Core.ILiteral {
  return { scalar: { generic } };
}

const values = {
  stringField: 'aString',
  integerField: 123,
  floatField: 123.456,
  nullField: null,
  booleanTrueField: true,
  booleanFalseField: false,
};

const structValues: { [k in keyof typeof values]: Protobuf.IValue } = {
  stringField: { stringValue: 'aString' },
  integerField: { numberValue: 123 },
  floatField: { numberValue: 123.456 },
  nullField: { nullValue: Protobuf.NullValue.NULL_VALUE },
  booleanTrueField: { boolValue: true },
  booleanFalseField: { boolValue: false },
};

type StructTestCase = [string, Core.ILiteral];
export const structTestCases: StructTestCase[] = [
  ['{}', structLiteral({ fields: {} })],
  // simple case with no lists or nested structs
  [stringifyValue({ ...values }), structLiteral({ fields: { ...structValues } })],
  // Nested struct value
  [
    stringifyValue({ nestedStruct: { ...values } }),
    structLiteral({
      fields: {
        nestedStruct: { structValue: { fields: { ...structValues } } },
      },
    }),
  ],
  // List
  [
    stringifyValue({ listField: Object.values(values) }),
    structLiteral({
      fields: {
        listField: {
          listValue: { values: Object.values(structValues) },
        },
      },
    }),
  ],
  // Nested struct with list
  [
    stringifyValue({ nestedStruct: { listField: Object.values(values) } }),
    structLiteral({
      fields: {
        nestedStruct: {
          structValue: {
            fields: {
              listField: {
                listValue: {
                  values: Object.values(structValues),
                },
              },
            },
          },
        },
      },
    }),
  ],
  // List with nested struct
  [
    stringifyValue({ listField: [{ ...values }] }),
    structLiteral({
      fields: {
        listField: {
          listValue: {
            values: [{ structValue: { fields: { ...structValues } } }],
          },
        },
      },
    }),
  ],
  // List with nested list
  [
    stringifyValue({ listField: [Object.values(values)] }),
    structLiteral({
      fields: {
        listField: {
          listValue: {
            values: [
              {
                listValue: {
                  values: Object.values(structValues),
                },
              },
            ],
          },
        },
      },
    }),
  ],
];
