import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import { InputType, InputValue } from '../../types';
import { literalNone } from '../constants';

// Defines type of value, input, and expected value of innermost `IScalar`
type PrimitiveTestParams = [InputType, any, Core.IPrimitive];

const validDateString = '2019-01-10T00:00:00.000Z'; // Dec 1, 2019

export const validityTestCases = {
    boolean: {
        invalid: ['randomString', {}, new Date()],
        valid: [true, 'true', 't', '1', 1, false, 'false', 'f', '0', 0]
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

export const literalTestCases: PrimitiveTestParams[] = [
    [InputType.Boolean, true, { boolean: true }],
    [InputType.Boolean, 'true', { boolean: true }],
    [InputType.Boolean, 't', { boolean: true }],
    [InputType.Boolean, '1', { boolean: true }],
    [InputType.Boolean, 1, { boolean: true }],
    [InputType.Boolean, false, { boolean: false }],
    [InputType.Boolean, 'false', { boolean: false }],
    [InputType.Boolean, 'f', { boolean: false }],
    [InputType.Boolean, '0', { boolean: false }],
    [InputType.Boolean, 0, { boolean: false }],
    [
        InputType.Datetime,
        new Date(validDateString),
        { datetime: dateToTimestamp(new Date(validDateString)) }
    ],
    [
        InputType.Datetime,
        validDateString,
        { datetime: dateToTimestamp(new Date(validDateString)) }
    ],
    [InputType.Duration, 0, { duration: millisecondsToDuration(0) }],
    [InputType.Duration, 10000, { duration: millisecondsToDuration(10000) }],
    [InputType.Float, 0, { floatValue: 0 }],
    [InputType.Float, '0', { floatValue: 0 }],
    [InputType.Float, -1.5, { floatValue: -1.5 }],
    [InputType.Float, '-1.5', { floatValue: -1.5 }],
    [InputType.Float, 1.5, { floatValue: 1.5 }],
    [InputType.Float, '1.5', { floatValue: 1.5 }],
    [InputType.Float, 1.25e10, { floatValue: 1.25e10 }],
    [InputType.Float, '1.25e10', { floatValue: 1.25e10 }],
    [InputType.Integer, 0, { integer: Long.fromNumber(0) }],
    [InputType.Integer, Long.fromNumber(0), { integer: Long.fromNumber(0) }],
    [InputType.Integer, '0', { integer: Long.fromNumber(0) }],
    [InputType.Integer, 1, { integer: Long.fromNumber(1) }],
    [InputType.Integer, Long.fromNumber(1), { integer: Long.fromNumber(1) }],
    [InputType.Integer, '1', { integer: Long.fromNumber(1) }],
    [InputType.Integer, -1, { integer: Long.fromNumber(-1) }],
    [InputType.Integer, Long.fromNumber(-1), { integer: Long.fromNumber(-1) }],
    [InputType.Integer, '-1', { integer: Long.fromNumber(-1) }],
    [InputType.Integer, Long.MAX_VALUE.toString(), { integer: Long.MAX_VALUE }],
    [InputType.Integer, Long.MAX_VALUE, { integer: Long.MAX_VALUE }],
    [InputType.Integer, Long.MIN_VALUE.toString(), { integer: Long.MIN_VALUE }],
    [InputType.Integer, Long.MIN_VALUE, { integer: Long.MIN_VALUE }],
    [InputType.String, '', { stringValue: '' }],
    [InputType.String, 'abcdefg', { stringValue: 'abcdefg' }]
];

function primitiveLiteral(primitive: Core.IPrimitive): Core.ILiteral {
    return { scalar: { primitive } };
}

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
        new Date(validDateString)
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
    [InputType.String, primitiveLiteral({ stringValue: 'abcdefg' }), 'abcdefg']
];
