import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Core } from 'flyteidl';
import * as Long from 'long';
import { inputToLiteral } from '../inputConverters';
import { InputProps, InputType } from '../types';

const baseInputProps: InputProps = {
    description: 'test',
    label: 'test',
    name: '',
    onChange: () => {},
    required: false,
    typeDefinition: { type: InputType.Unknown }
};

function makeSimpleInput(type: InputType, value: any): InputProps {
    return { ...baseInputProps, value, typeDefinition: { type } };
}

function makeCollectionInput(type: InputType, value: string): InputProps {
    return {
        ...baseInputProps,
        value,
        typeDefinition: { type: InputType.Collection, subtype: { type } }
    };
}

function makeNestedCollectionInput(type: InputType, value: string): InputProps {
    return {
        ...baseInputProps,
        value,
        typeDefinition: {
            type: InputType.Collection,
            subtype: { type: InputType.Collection, subtype: { type } }
        }
    };
}

// Defines type of value, input, and expected value of innermost `IScalar`
type PrimitiveTestParams = [InputType, any, Core.IPrimitive];

const validDateString = '2019-01-10T00:00:00.000Z'; // Dec 1, 2019

describe('inputToLiteral', () => {
    describe('SimpleInput', () => {
        const simpleTestCases: PrimitiveTestParams[] = [
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
                new Date('2019-01-10T00:00:00.000Z'),
                { datetime: dateToTimestamp(new Date(validDateString)) }
            ],
            [
                InputType.Datetime,
                validDateString,
                { datetime: dateToTimestamp(new Date(validDateString)) }
            ],
            [InputType.Duration, 0, { duration: millisecondsToDuration(0) }],
            [
                InputType.Duration,
                10000,
                { duration: millisecondsToDuration(10000) }
            ],
            [InputType.Float, 0, { floatValue: 0 }],
            [InputType.Float, '0', { floatValue: 0 }],
            [InputType.Float, -1.5, { floatValue: -1.5 }],
            [InputType.Float, '-1.5', { floatValue: -1.5 }],
            [InputType.Float, 1.5, { floatValue: 1.5 }],
            [InputType.Float, '1.5', { floatValue: 1.5 }],
            [InputType.Float, 1.25e10, { floatValue: 1.25e10 }],
            [InputType.Float, '1.25e10', { floatValue: 1.25e10 }],
            [InputType.Integer, 0, { integer: Long.fromNumber(0) }],
            [
                InputType.Integer,
                Long.fromNumber(0),
                { integer: Long.fromNumber(0) }
            ],
            [InputType.Integer, '0', { integer: Long.fromNumber(0) }],
            [InputType.Integer, 1, { integer: Long.fromNumber(1) }],
            [
                InputType.Integer,
                Long.fromNumber(1),
                { integer: Long.fromNumber(1) }
            ],
            [InputType.Integer, '1', { integer: Long.fromNumber(1) }],
            [InputType.Integer, -1, { integer: Long.fromNumber(-1) }],
            [
                InputType.Integer,
                Long.fromNumber(-1),
                { integer: Long.fromNumber(-1) }
            ],
            [InputType.Integer, '-1', { integer: Long.fromNumber(-1) }],
            [
                InputType.Integer,
                Long.MAX_VALUE.toString(),
                { integer: Long.MAX_VALUE }
            ],
            [InputType.Integer, Long.MAX_VALUE, { integer: Long.MAX_VALUE }],
            [
                InputType.Integer,
                Long.MIN_VALUE.toString(),
                { integer: Long.MIN_VALUE }
            ],
            [InputType.Integer, Long.MIN_VALUE, { integer: Long.MIN_VALUE }],
            [InputType.String, '', { stringValue: '' }],
            [InputType.String, 'abcdefg', { stringValue: 'abcdefg' }]
        ];

        simpleTestCases.map(([type, input, output]) =>
            it(`Should correctly convert ${type}: ${input} (${typeof input})`, () => {
                const result = inputToLiteral(makeSimpleInput(type, input));
                expect(result.scalar!.primitive).toEqual(output);
            })
        );

        simpleTestCases.map(([type, input, output]) => {
            let value: any;
            if (['boolean', 'number'].includes(typeof input)) {
                value = input;
            } else if (input instanceof Date) {
                value = `"${input.toISOString()}"`;
            } else {
                value = `"${input}"`;
            }

            it(`Should correctly convert collection of type ${type}: [${value}] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeCollectionInput(type, `[${value}]`)
                );
                expect(
                    result.collection!.literals![0].scalar!.primitive
                ).toEqual(output);
            });

            it(`Should correctly convert nested collection of type ${type}: [[${value}]] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeNestedCollectionInput(type, `[[${value}]]`)
                );
                expect(
                    result.collection!.literals![0].collection!.literals![0]
                        .scalar!.primitive
                ).toEqual(output);
            });
        });

        // Unsupported / None types
        [
            InputType.Binary,
            InputType.Blob,
            InputType.Error,
            InputType.Map,
            InputType.None,
            InputType.Schema,
            InputType.Struct,
            InputType.Unknown
        ].map(type =>
            it(`Should return empty value for type: ${type}`, () => {
                expect(
                    inputToLiteral(makeSimpleInput(type, '')).scalar
                ).toEqual({ noneType: {} });
            })
        );
    });
    describe('Collections', () => {});
    describe('NestedCollections', () => {});
});
