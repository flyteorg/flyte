import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import { BlobDimensionality } from 'models';
import { blobLiteral, primitiveLiteral } from '../../__mocks__/utils';
import { InputType, InputValue } from '../../types';
import { literalNone } from '../constants';

// Defines type of value, input, and expected value of a `Core.ILiteral`
type LiteralTestParams = [InputType, any, Core.ILiteral];

const validDateString = '2019-01-10T00:00:00.000Z'; // Dec 1, 2019

export const supportedPrimitives = [
    InputType.Boolean,
    InputType.Blob,
    InputType.Datetime,
    InputType.Duration,
    InputType.Float,
    InputType.Integer
];

export const unsupportedTypes = [
    InputType.Binary,
    InputType.Error,
    InputType.Map,
    InputType.None,
    InputType.Schema,
    InputType.Struct
];

export const validityTestCases = {
    boolean: {
        invalid: ['randomString', {}, new Date()],
        valid: [true, 'true', 't', '1', 1, false, 'false', 'f', '0', 0]
    },
    blob: {
        invalid: [
            {
                uri: 5,
                format: 'csv',
                dimensionality: BlobDimensionality.SINGLE
            },
            {
                uri: 'path',
                format: 5,
                dimensionality: BlobDimensionality.SINGLE
            },
            { uri: 'path', dimensionality: 'notAnEnumValue' },
            { uri: 'path', dimensionality: 1000 },
            { uri: 'path' }
        ],
        valid: [
            {
                uri: 'path',
                format: 'csv',
                dimensionality: BlobDimensionality.SINGLE
            },
            {
                uri: 'path',
                format: 'csv',
                dimensionality: BlobDimensionality.MULTIPART
            },
            { uri: 'path', format: 'csv', dimensionality: 'single' },
            { uri: 'path', format: 'csv', dimensionality: 'SINGLE' },
            { uri: 'path', format: 'csv', dimensionality: 'multipart' },
            { uri: 'path', format: 'csv', dimensionality: 'MULTIPART' },
            {
                uri: 'path',
                format: '',
                dimensionality: BlobDimensionality.SINGLE
            },
            { uri: 'path', dimensionality: BlobDimensionality.SINGLE }
        ]
    },
    datetime: {
        invalid: ['abc', true],
        valid: [new Date(validDateString), validDateString]
    },
    duration: {
        // Same cases used for float
        invalid: ['abc', {}, true, new Date()],
        valid: [0, '0', -1.5, '-1.5', 1.5, '1.5', 1.25e10, '1.25e10']
    },
    float: {
        invalid: ['abc', {}, true, new Date()],
        valid: [0, '0', -1.5, '-1.5', 1.5, '1.5', 1.25e10, '1.25e10']
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
            Long.MIN_VALUE
        ]
    },
    string: { invalid: [123, true, new Date(), {}], valid: ['', 'abcdefg'] }
};

export const literalTestCases: LiteralTestParams[] = [
    [InputType.Boolean, true, primitiveLiteral({ boolean: true })],
    [InputType.Boolean, 'true', primitiveLiteral({ boolean: true })],
    [InputType.Boolean, 't', primitiveLiteral({ boolean: true })],
    [InputType.Boolean, '1', primitiveLiteral({ boolean: true })],
    [InputType.Boolean, 1, primitiveLiteral({ boolean: true })],
    [InputType.Boolean, false, primitiveLiteral({ boolean: false })],
    [InputType.Boolean, 'false', primitiveLiteral({ boolean: false })],
    [InputType.Boolean, 'f', primitiveLiteral({ boolean: false })],
    [InputType.Boolean, '0', primitiveLiteral({ boolean: false })],
    [InputType.Boolean, 0, primitiveLiteral({ boolean: false })],
    [
        InputType.Datetime,
        new Date(validDateString),
        primitiveLiteral({
            datetime: dateToTimestamp(new Date(validDateString))
        })
    ],
    [
        InputType.Datetime,
        validDateString,
        primitiveLiteral({
            datetime: dateToTimestamp(new Date(validDateString))
        })
    ],
    [
        InputType.Duration,
        0,
        primitiveLiteral({ duration: millisecondsToDuration(0) })
    ],
    [
        InputType.Duration,
        10000,
        primitiveLiteral({ duration: millisecondsToDuration(10000) })
    ],
    [InputType.Float, 0, primitiveLiteral({ floatValue: 0 })],
    [InputType.Float, '0', primitiveLiteral({ floatValue: 0 })],
    [InputType.Float, -1.5, primitiveLiteral({ floatValue: -1.5 })],
    [InputType.Float, '-1.5', primitiveLiteral({ floatValue: -1.5 })],
    [InputType.Float, 1.5, primitiveLiteral({ floatValue: 1.5 })],
    [InputType.Float, '1.5', primitiveLiteral({ floatValue: 1.5 })],
    [InputType.Float, 1.25e10, primitiveLiteral({ floatValue: 1.25e10 })],
    [InputType.Float, '1.25e10', primitiveLiteral({ floatValue: 1.25e10 })],
    [InputType.Integer, 0, primitiveLiteral({ integer: Long.fromNumber(0) })],
    [
        InputType.Integer,
        Long.fromNumber(0),
        primitiveLiteral({ integer: Long.fromNumber(0) })
    ],
    [InputType.Integer, '0', primitiveLiteral({ integer: Long.fromNumber(0) })],
    [InputType.Integer, 1, primitiveLiteral({ integer: Long.fromNumber(1) })],
    [
        InputType.Integer,
        Long.fromNumber(1),
        primitiveLiteral({ integer: Long.fromNumber(1) })
    ],
    [InputType.Integer, '1', primitiveLiteral({ integer: Long.fromNumber(1) })],
    [InputType.Integer, -1, primitiveLiteral({ integer: Long.fromNumber(-1) })],
    [
        InputType.Integer,
        Long.fromNumber(-1),
        primitiveLiteral({ integer: Long.fromNumber(-1) })
    ],
    [
        InputType.Integer,
        '-1',
        primitiveLiteral({ integer: Long.fromNumber(-1) })
    ],
    [
        InputType.Integer,
        Long.MAX_VALUE.toString(),
        primitiveLiteral({ integer: Long.MAX_VALUE })
    ],
    [
        InputType.Integer,
        Long.MAX_VALUE,
        primitiveLiteral({ integer: Long.MAX_VALUE })
    ],
    [
        InputType.Integer,
        Long.MIN_VALUE.toString(),
        primitiveLiteral({ integer: Long.MIN_VALUE })
    ],
    [
        InputType.Integer,
        Long.MIN_VALUE,
        primitiveLiteral({ integer: Long.MIN_VALUE })
    ],
    [InputType.String, '', primitiveLiteral({ stringValue: '' })],
    [InputType.String, 'abcdefg', primitiveLiteral({ stringValue: 'abcdefg' })],
    // Standard Blob
    [
        InputType.Blob,
        {
            uri: 's3://somePath',
            format: 'csv',
            dimensionality: BlobDimensionality.SINGLE
        },
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            format: 'csv',
            uri: 's3://somePath'
        })
    ],
    // Multi-part blob
    [
        InputType.Blob,
        {
            dimensionality: BlobDimensionality.MULTIPART,
            format: 'csv',
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.MULTIPART,
            format: 'csv',
            uri: 's3://somePath'
        })
    ],
    // Blob with missing format
    [
        InputType.Blob,
        {
            dimensionality: BlobDimensionality.SINGLE,
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            uri: 's3://somePath'
        })
    ],
    // Blob with empty format string
    [
        InputType.Blob,
        {
            dimensionality: BlobDimensionality.SINGLE,
            format: '',
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            uri: 's3://somePath'
        })
    ],
    // Blobs using lowercase string for dimensionality
    [
        InputType.Blob,
        {
            dimensionality: 'single',
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            uri: 's3://somePath'
        })
    ],
    [
        InputType.Blob,
        {
            dimensionality: 'multipart',
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.MULTIPART,
            uri: 's3://somePath'
        })
    ],
    // Blobs using uppercase string for dimensionality
    [
        InputType.Blob,
        {
            dimensionality: 'SINGLE',
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            uri: 's3://somePath'
        })
    ],
    [
        InputType.Blob,
        {
            dimensionality: 'MULTIPART',
            uri: 's3://somePath'
        },
        blobLiteral({
            dimensionality: BlobDimensionality.MULTIPART,
            uri: 's3://somePath'
        })
    ],
    // Blob missing URI (results in None)
    [
        InputType.Blob,
        {
            format: 'csv',
            dimensionality: 'MULTIPART'
        },
        literalNone()
    ],
    // Blob which is not an object (results in None)
    [InputType.Blob, undefined, literalNone()]
];

type InputToLiteralTestParams = [
    InputType,
    Core.ILiteral,
    InputValue | undefined
];
export const literalToInputTestCases: InputToLiteralTestParams[] = [
    [InputType.Boolean, primitiveLiteral({ boolean: true }), true],
    [InputType.Boolean, primitiveLiteral({ boolean: false }), false],
    [
        InputType.Datetime,
        primitiveLiteral({
            datetime: dateToTimestamp(new Date(validDateString))
        }),
        validDateString
    ],
    [
        InputType.Duration,
        primitiveLiteral({ duration: millisecondsToDuration(0) }),
        0
    ],
    [
        InputType.Duration,
        primitiveLiteral({ duration: millisecondsToDuration(10000) }),
        10000
    ],
    [
        InputType.Duration,
        primitiveLiteral({ duration: millisecondsToDuration(1.5) }),
        1.5
    ],
    [InputType.Float, primitiveLiteral({ floatValue: 0 }), 0],
    [InputType.Float, primitiveLiteral({ floatValue: -1.5 }), -1.5],
    [InputType.Float, primitiveLiteral({ floatValue: 1.5 }), 1.5],
    [InputType.Float, primitiveLiteral({ floatValue: 1.25e10 }), 1.25e10],
    // Integers will be returned as strings because they may overflow numbers
    [InputType.Integer, primitiveLiteral({ integer: Long.fromNumber(0) }), '0'],
    [InputType.Integer, primitiveLiteral({ integer: Long.fromNumber(1) }), '1'],
    [
        InputType.Integer,
        primitiveLiteral({ integer: Long.fromNumber(-1) }),
        '-1'
    ],
    [
        InputType.Integer,
        primitiveLiteral({ integer: Long.MAX_VALUE }),
        Long.MAX_VALUE.toString()
    ],
    [
        InputType.Integer,
        primitiveLiteral({ integer: Long.MIN_VALUE }),
        Long.MIN_VALUE.toString()
    ],
    [InputType.String, primitiveLiteral({ stringValue: '' }), ''],
    [InputType.String, primitiveLiteral({ stringValue: 'abcdefg' }), 'abcdefg'],
    // Standard Blob case
    [
        InputType.Blob,
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            format: 'csv',
            uri: 's3://somePath'
        }),
        {
            dimensionality: BlobDimensionality.SINGLE,
            format: 'csv',
            uri: 's3://somePath'
        }
    ],
    // Multipart blob
    [
        InputType.Blob,
        blobLiteral({
            dimensionality: BlobDimensionality.MULTIPART,
            format: 'csv',
            uri: 's3://somePath'
        }),
        {
            dimensionality: BlobDimensionality.MULTIPART,
            format: 'csv',
            uri: 's3://somePath'
        }
    ],
    // Empty uri
    [
        InputType.Blob,
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            format: 'csv'
        }),
        {
            dimensionality: BlobDimensionality.SINGLE,
            format: 'csv',
            uri: ''
        }
    ],
    // Empty format string
    [
        InputType.Blob,
        blobLiteral({
            dimensionality: BlobDimensionality.SINGLE,
            format: '',
            uri: 's3://somePath'
        }),
        {
            dimensionality: BlobDimensionality.SINGLE,
            uri: 's3://somePath'
        }
    ],
    // Missing dimensionality
    [
        InputType.Blob,
        blobLiteral({
            format: 'csv',
            uri: 's3://somePath'
        }),
        {
            dimensionality: BlobDimensionality.SINGLE,
            format: 'csv',
            uri: 's3://somePath'
        }
    ]
];
