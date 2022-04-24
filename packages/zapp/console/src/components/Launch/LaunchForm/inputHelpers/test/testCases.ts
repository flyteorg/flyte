import { dateToTimestamp, millisecondsToDuration, stringifyValue } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import { BlobDimensionality, SchemaColumnType } from 'models/Common/types';
import { InputType, InputTypeDefinition, InputValue } from '../../types';
import { blobLiteral, primitiveLiteral } from '../../__mocks__/utils';
import { literalNone } from '../constants';
import { structTestCases } from './structTestCases';

// Defines type of value, input, and expected value of a `Core.ILiteral`
type InputToLiteralTestParams = [InputTypeDefinition, any, Core.ILiteral];

type InputTypeKey =
  | 'binary'
  | 'boolean'
  | 'blobSingle'
  | 'blobMulti'
  | 'datetime'
  | 'duration'
  | 'error'
  | 'integer'
  | 'float'
  | 'map'
  | 'none'
  | 'schema'
  | 'string'
  | 'struct'
  | 'unknown';
export const inputTypes: Record<InputTypeKey, InputTypeDefinition> = {
  binary: {
    literalType: { simple: Core.SimpleType.BINARY },
    type: InputType.Binary,
  },
  boolean: {
    literalType: { simple: Core.SimpleType.BOOLEAN },
    type: InputType.Boolean,
  },
  blobSingle: {
    literalType: { blob: { dimensionality: BlobDimensionality.SINGLE } },
    type: InputType.Blob,
  },
  blobMulti: {
    literalType: { blob: { dimensionality: BlobDimensionality.MULTIPART } },
    type: InputType.Blob,
  },
  datetime: {
    literalType: { simple: Core.SimpleType.DATETIME },
    type: InputType.Datetime,
  },
  duration: {
    literalType: { simple: Core.SimpleType.DURATION },
    type: InputType.Duration,
  },
  error: {
    literalType: { simple: Core.SimpleType.ERROR },
    type: InputType.Error,
  },
  integer: {
    literalType: { simple: Core.SimpleType.INTEGER },
    type: InputType.Integer,
  },
  float: {
    literalType: { simple: Core.SimpleType.FLOAT },
    type: InputType.Float,
  },
  map: {
    literalType: {
      mapValueType: { simple: Core.SimpleType.STRING },
    },
    type: InputType.Map,
  },
  none: {
    literalType: { simple: Core.SimpleType.NONE },
    type: InputType.None,
  },
  schema: {
    literalType: {
      schema: {
        columns: [{ name: 'column1', type: SchemaColumnType.STRING }],
      },
    },
    type: InputType.Schema,
  },
  string: {
    literalType: { simple: Core.SimpleType.STRING },
    type: InputType.String,
  },
  struct: {
    literalType: { simple: Core.SimpleType.STRUCT },
    type: InputType.Struct,
  },
  unknown: {
    literalType: { simple: Core.SimpleType.NONE },
    type: InputType.Unknown,
  },
};

const validDateString = '2019-01-10T00:00:00.000Z'; // Dec 1, 2019

export const supportedPrimitives: InputTypeDefinition[] = [
  inputTypes.boolean,
  inputTypes.blobSingle,
  inputTypes.blobMulti,
  inputTypes.datetime,
  inputTypes.duration,
  inputTypes.float,
  inputTypes.integer,
  inputTypes.schema,
  inputTypes.struct,
];

export const unsupportedTypes: InputTypeDefinition[] = [
  inputTypes.binary,
  inputTypes.error,
  inputTypes.map,
  inputTypes.none,
];

export const validityTestCases = {
  boolean: {
    invalid: ['randomString', {}, new Date()],
    valid: [true, 'true', 't', '1', 1, false, 'false', 'f', '0', 0],
  },
  blob: {
    invalid: [
      {
        uri: 5,
        format: 'csv',
        dimensionality: BlobDimensionality.SINGLE,
      },
      {
        uri: 'path',
        format: 5,
        dimensionality: BlobDimensionality.SINGLE,
      },
      { uri: 'path', dimensionality: 'notAnEnumValue' },
      { uri: 'path', dimensionality: 1000 },
      { uri: 'path' },
    ],
    valid: [
      {
        uri: 'path',
        format: 'csv',
        dimensionality: BlobDimensionality.SINGLE,
      },
      {
        uri: 'path',
        format: 'csv',
        dimensionality: BlobDimensionality.MULTIPART,
      },
      { uri: 'path', format: 'csv', dimensionality: 'single' },
      { uri: 'path', format: 'csv', dimensionality: 'SINGLE' },
      { uri: 'path', format: 'csv', dimensionality: 'multipart' },
      { uri: 'path', format: 'csv', dimensionality: 'MULTIPART' },
      {
        uri: 'path',
        format: '',
        dimensionality: BlobDimensionality.SINGLE,
      },
      { uri: 'path', dimensionality: BlobDimensionality.SINGLE },
    ],
  },
  datetime: {
    invalid: ['abc', true],
    valid: [new Date(validDateString), validDateString],
  },
  duration: {
    // Same cases used for float
    invalid: ['abc', {}, true, new Date()],
    valid: [0, '0', -1.5, '-1.5', 1.5, '1.5', 1.25e10, '1.25e10'],
  },
  float: {
    invalid: ['abc', {}, true, new Date()],
    valid: [0, '0', -1.5, '-1.5', 1.5, '1.5', 1.25e10, '1.25e10'],
  },
  integer: {
    invalid: ['a', {}, true, new Date(), 1.1, 0 / 0, '1.1', '1a'],
    valid: [
      0,
      0.0,
      Long.fromNumber(0),
      '0',
      1,
      1.0,
      Long.fromNumber(1),
      '1',
      -1,
      -1.0,
      Long.fromNumber(-1),
      '-1',
      Long.MAX_VALUE.toString(),
      Long.MAX_VALUE,
      Long.MIN_VALUE.toString(),
      Long.MIN_VALUE,
    ],
  },
  // schema is just a specialized string input, so it has the same validity cases as string
  schema: { invalid: [123, true, new Date(), {}], valid: ['', 'abcdefg'] },
  string: { invalid: [123, true, new Date(), {}], valid: ['', 'abcdefg'] },
  struct: {
    invalid: [123, true, new Date(), {}, 'nonObjectString', '[]', '{garbageobject}'],
    valid: [
      // Valid use case is any stringified POJO
      stringifyValue({
        someString: 'someValue',
        someNumber: 123,
        someBoolean: true,
        nestedStruct: { someOtherString: 'someOtherValue' },
        nestedList: [123, 'stringListValue'],
        nullValue: null,
      }),
    ],
  },
};

/** Test cases for converting a *valid* input value to its corresponding ILiteral
 * representation. */
export const literalTestCases: InputToLiteralTestParams[] = [
  [inputTypes.boolean, true, primitiveLiteral({ boolean: true })],
  [inputTypes.boolean, 'true', primitiveLiteral({ boolean: true })],
  [inputTypes.boolean, 't', primitiveLiteral({ boolean: true })],
  [inputTypes.boolean, '1', primitiveLiteral({ boolean: true })],
  [inputTypes.boolean, 1, primitiveLiteral({ boolean: true })],
  [inputTypes.boolean, false, primitiveLiteral({ boolean: false })],
  [inputTypes.boolean, 'false', primitiveLiteral({ boolean: false })],
  [inputTypes.boolean, 'f', primitiveLiteral({ boolean: false })],
  [inputTypes.boolean, '0', primitiveLiteral({ boolean: false })],
  [inputTypes.boolean, 0, primitiveLiteral({ boolean: false })],
  [
    inputTypes.datetime,
    new Date(validDateString),
    primitiveLiteral({
      datetime: dateToTimestamp(new Date(validDateString)),
    }),
  ],
  [
    inputTypes.datetime,
    validDateString,
    primitiveLiteral({
      datetime: dateToTimestamp(new Date(validDateString)),
    }),
  ],
  [inputTypes.duration, 0, primitiveLiteral({ duration: millisecondsToDuration(0) })],
  [inputTypes.duration, 10000, primitiveLiteral({ duration: millisecondsToDuration(10000) })],
  [inputTypes.float, 0, primitiveLiteral({ floatValue: 0 })],
  [inputTypes.float, '0', primitiveLiteral({ floatValue: 0 })],
  [inputTypes.float, -1.5, primitiveLiteral({ floatValue: -1.5 })],
  [inputTypes.float, '-1.5', primitiveLiteral({ floatValue: -1.5 })],
  [inputTypes.float, 1.5, primitiveLiteral({ floatValue: 1.5 })],
  [inputTypes.float, '1.5', primitiveLiteral({ floatValue: 1.5 })],
  [inputTypes.float, 1.25e10, primitiveLiteral({ floatValue: 1.25e10 })],
  [inputTypes.float, '1.25e10', primitiveLiteral({ floatValue: 1.25e10 })],
  [inputTypes.integer, 0, primitiveLiteral({ integer: Long.fromNumber(0) })],
  [inputTypes.integer, Long.fromNumber(0), primitiveLiteral({ integer: Long.fromNumber(0) })],
  [inputTypes.integer, '0', primitiveLiteral({ integer: Long.fromNumber(0) })],
  [inputTypes.integer, 1, primitiveLiteral({ integer: Long.fromNumber(1) })],
  [inputTypes.integer, Long.fromNumber(1), primitiveLiteral({ integer: Long.fromNumber(1) })],
  [inputTypes.integer, '1', primitiveLiteral({ integer: Long.fromNumber(1) })],
  [inputTypes.integer, -1, primitiveLiteral({ integer: Long.fromNumber(-1) })],
  [inputTypes.integer, Long.fromNumber(-1), primitiveLiteral({ integer: Long.fromNumber(-1) })],
  [inputTypes.integer, '-1', primitiveLiteral({ integer: Long.fromNumber(-1) })],
  [inputTypes.integer, Long.MAX_VALUE.toString(), primitiveLiteral({ integer: Long.MAX_VALUE })],
  [inputTypes.integer, Long.MAX_VALUE, primitiveLiteral({ integer: Long.MAX_VALUE })],
  [inputTypes.integer, Long.MIN_VALUE.toString(), primitiveLiteral({ integer: Long.MIN_VALUE })],
  [inputTypes.integer, Long.MIN_VALUE, primitiveLiteral({ integer: Long.MIN_VALUE })],
  [
    inputTypes.schema,
    '',
    {
      scalar: {
        schema: { type: inputTypes.schema.literalType.schema, uri: '' },
      },
    },
  ],
  [
    inputTypes.schema,
    's3://someUri',
    {
      scalar: {
        schema: {
          type: inputTypes.schema.literalType.schema,
          uri: 's3://someUri',
        },
      },
    },
  ],
  [inputTypes.string, '', primitiveLiteral({ stringValue: '' })],
  [inputTypes.string, 'abcdefg', primitiveLiteral({ stringValue: 'abcdefg' })],
  // Standard Blob
  [
    inputTypes.blobSingle,
    {
      uri: 's3://somePath',
      format: 'csv',
      dimensionality: BlobDimensionality.SINGLE,
    },
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      format: 'csv',
      uri: 's3://somePath',
    }),
  ],
  // Multi-part blob
  [
    inputTypes.blobSingle,
    {
      dimensionality: BlobDimensionality.MULTIPART,
      format: 'csv',
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.MULTIPART,
      format: 'csv',
      uri: 's3://somePath',
    }),
  ],
  // Blob with missing format
  [
    inputTypes.blobSingle,
    {
      dimensionality: BlobDimensionality.SINGLE,
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      uri: 's3://somePath',
    }),
  ],
  // Blob with empty format string
  [
    inputTypes.blobSingle,
    {
      dimensionality: BlobDimensionality.SINGLE,
      format: '',
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      uri: 's3://somePath',
    }),
  ],
  // Blobs using lowercase string for dimensionality
  [
    inputTypes.blobSingle,
    {
      dimensionality: 'single',
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      uri: 's3://somePath',
    }),
  ],
  [
    inputTypes.blobMulti,
    {
      dimensionality: 'multipart',
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.MULTIPART,
      uri: 's3://somePath',
    }),
  ],
  // Blobs using uppercase string for dimensionality
  [
    inputTypes.blobMulti,
    {
      dimensionality: 'SINGLE',
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      uri: 's3://somePath',
    }),
  ],
  [
    inputTypes.blobMulti,
    {
      dimensionality: 'MULTIPART',
      uri: 's3://somePath',
    },
    blobLiteral({
      dimensionality: BlobDimensionality.MULTIPART,
      uri: 's3://somePath',
    }),
  ],
  // Blob missing URI (results in None)
  [
    inputTypes.blobMulti,
    {
      format: 'csv',
      dimensionality: 'MULTIPART',
    },
    literalNone(),
  ],
  // Blob which is not an object (results in None)
  [inputTypes.blobMulti, undefined, literalNone()],
  ...structTestCases.map<InputToLiteralTestParams>(([stringValue, literalValue]) => [
    inputTypes.struct,
    stringValue,
    literalValue,
  ]),
];

type LiteralToInputTestParams = [InputTypeDefinition, Core.ILiteral, InputValue | undefined];

/** Test cases for converting a Core.ILiteral to a usable InputValue */
export const literalToInputTestCases: LiteralToInputTestParams[] = [
  [inputTypes.boolean, primitiveLiteral({ boolean: true }), true],
  [inputTypes.boolean, primitiveLiteral({ boolean: false }), false],
  [
    inputTypes.datetime,
    primitiveLiteral({
      datetime: dateToTimestamp(new Date(validDateString)),
    }),
    validDateString,
  ],
  [inputTypes.duration, primitiveLiteral({ duration: millisecondsToDuration(0) }), 0],
  [inputTypes.duration, primitiveLiteral({ duration: millisecondsToDuration(10000) }), 10000],
  [inputTypes.duration, primitiveLiteral({ duration: millisecondsToDuration(1.5) }), 1.5],
  [inputTypes.float, primitiveLiteral({ floatValue: 0 }), 0],
  [inputTypes.float, primitiveLiteral({ floatValue: -1.5 }), -1.5],
  [inputTypes.float, primitiveLiteral({ floatValue: 1.5 }), 1.5],
  [inputTypes.float, primitiveLiteral({ floatValue: 1.25e10 }), 1.25e10],
  // Integers will be returned as strings because they may overflow numbers
  [inputTypes.integer, primitiveLiteral({ integer: Long.fromNumber(0) }), '0'],
  [inputTypes.integer, primitiveLiteral({ integer: Long.fromNumber(1) }), '1'],
  [inputTypes.integer, primitiveLiteral({ integer: Long.fromNumber(-1) }), '-1'],
  [inputTypes.integer, primitiveLiteral({ integer: Long.MAX_VALUE }), Long.MAX_VALUE.toString()],
  [inputTypes.integer, primitiveLiteral({ integer: Long.MIN_VALUE }), Long.MIN_VALUE.toString()],
  [inputTypes.schema, { scalar: { schema: { uri: '' } } }, ''],
  [inputTypes.schema, { scalar: { schema: { uri: 's3://someUri' } } }, 's3://someUri'],
  [inputTypes.string, primitiveLiteral({ stringValue: '' }), ''],
  [inputTypes.string, primitiveLiteral({ stringValue: 'abcdefg' }), 'abcdefg'],
  // Standard Blob case
  [
    inputTypes.blobSingle,
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      format: 'csv',
      uri: 's3://somePath',
    }),
    {
      dimensionality: BlobDimensionality.SINGLE,
      format: 'csv',
      uri: 's3://somePath',
    },
  ],
  // Multipart blob
  [
    inputTypes.blobMulti,
    blobLiteral({
      dimensionality: BlobDimensionality.MULTIPART,
      format: 'csv',
      uri: 's3://somePath',
    }),
    {
      dimensionality: BlobDimensionality.MULTIPART,
      format: 'csv',
      uri: 's3://somePath',
    },
  ],
  // Empty uri
  [
    inputTypes.blobSingle,
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      format: 'csv',
    }),
    {
      dimensionality: BlobDimensionality.SINGLE,
      format: 'csv',
      uri: '',
    },
  ],
  // Empty format string
  [
    inputTypes.blobSingle,
    blobLiteral({
      dimensionality: BlobDimensionality.SINGLE,
      format: '',
      uri: 's3://somePath',
    }),
    {
      dimensionality: BlobDimensionality.SINGLE,
      uri: 's3://somePath',
    },
  ],
  // Missing dimensionality
  [
    inputTypes.blobSingle,
    blobLiteral({
      format: 'csv',
      uri: 's3://somePath',
    }),
    {
      dimensionality: BlobDimensionality.SINGLE,
      format: 'csv',
      uri: 's3://somePath',
    },
  ],
  ...structTestCases.map<LiteralToInputTestParams>(([stringValue, literalValue]) => [
    inputTypes.struct,
    literalValue,
    stringValue,
  ]),
];
