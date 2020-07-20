import { Core } from 'flyteidl';
import * as Long from 'long';
import { BlobDimensionality } from 'models';
import { primitiveLiteral } from '../../__mocks__/utils';
import { InputProps, InputType } from '../../types';
import { literalNone } from '../constants';
import { getHelperForInput } from '../getHelperForInput';
import {
    inputToLiteral,
    literalToInputValue,
    validateInput
} from '../inputHelpers';
import { collectionChildToString } from '../utils';
import {
    literalTestCases,
    literalToInputTestCases,
    supportedPrimitives,
    unsupportedTypes,
    validityTestCases
} from './testCases';

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

describe('literalToInputValue', () => {
    describe('Primitives', () => {
        literalToInputTestCases.map(([type, input, output]) =>
            it(`should correctly convert ${type}: ${JSON.stringify(
                input
            )}`, () =>
                expect(literalToInputValue({ type }, input)).toEqual(output))
        );

        supportedPrimitives.map(type =>
            it(`should convert None value for ${type} to undefined`, () =>
                expect(
                    literalToInputValue({ type }, literalNone())
                ).toBeUndefined())
        );

        it('should correctly convert noneType to undefined', () =>
            expect(
                literalToInputValue({ type: InputType.None }, literalNone())
            ).toEqual(undefined));
    });

    describe('Collections', () => {
        literalToInputTestCases.map(([type, input, output]) => {
            it(`should correctly convert collection of ${type}: ${JSON.stringify(
                input
            )}`, () => {
                const collection: Core.ILiteral = {
                    collection: {
                        // Duplicate it to test comma separation
                        literals: [input, input]
                    }
                };
                const stringifiedValue = collectionChildToString(type, output);
                const expectedString = `[${stringifiedValue},${stringifiedValue}]`;
                const result = literalToInputValue(
                    { type: InputType.Collection, subtype: { type } },
                    collection
                );
                expect(result).toEqual(expectedString);
            });
        });

        it('should return empty for noneType literals', () => {
            const collection: Core.ILiteral = {
                collection: {
                    // Duplicate it to test comma separation
                    literals: [literalNone(), literalNone()]
                }
            };

            expect(
                literalToInputValue(
                    {
                        type: InputType.Collection,
                        subtype: { type: InputType.None }
                    },
                    collection
                )
            ).toEqual('[]');
        });
    });

    it('should return system default if parsing literal fails', () => {
        const { defaultValue } = getHelperForInput(InputType.Boolean);
        expect(
            literalToInputValue(
                { type: InputType.Boolean },
                // Invalid boolean input value because it uses the string field
                { scalar: { primitive: { stringValue: 'whoops' } } }
            )
        ).toEqual(defaultValue);
    });
});

describe('inputToLiteral', () => {
    describe('Scalars', () => {
        literalTestCases.map(([type, input, output]) => {
            it(`should correctly convert ${type}: ${JSON.stringify(
                input
            )} (${typeof input})`, () =>
                expect(inputToLiteral(makeSimpleInput(type, input))).toEqual(
                    output
                ));
        });
    });

    describe('Collections', () => {
        literalTestCases.map(([type, input, output]) => {
            let value: any;
            if (['boolean', 'number'].includes(typeof input)) {
                value = input;
            } else if (input == null) {
                value = 'null';
            } else if (typeof input === 'string' || Long.isLong(input)) {
                value = `"${input}"`;
            } else if (input instanceof Date) {
                value = `"${input.toISOString()}"`;
            } else {
                value = JSON.stringify(input);
            }

            it(`should correctly convert collection of type ${type}: [${JSON.stringify(
                value
            )}] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeCollectionInput(type, `[${value}]`)
                );
                expect(result.collection!.literals![0]).toEqual(output);
            });

            it(`should correctly convert nested collection of type ${type}: [[${JSON.stringify(
                value
            )}]] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeNestedCollectionInput(type, `[[${value}]]`)
                );
                expect(
                    result.collection!.literals![0].collection!.literals![0]
                ).toEqual(output);
            });
        });
    });

    describe('Unsupported Types', () => {
        unsupportedTypes.map(type =>
            it(`should return empty value for type: ${type}`, () => {
                expect(
                    inputToLiteral(makeSimpleInput(type, '')).scalar
                ).toEqual({ noneType: {} });
            })
        );
    });

    it('Should return initial value for inputs with no value', () => {
        const simpleInput = makeSimpleInput(
            InputType.String,
            primitiveLiteral({ stringValue: '' })
        );
        const initialValue = primitiveLiteral({ stringValue: 'abcdefg' });
        simpleInput.required = true;
        simpleInput.initialValue = initialValue;
        delete simpleInput.value;
        expect(inputToLiteral(simpleInput)).toEqual(initialValue);
    });
});

function generateValidityTests(
    type: InputType,
    { valid, invalid }: { valid: any[]; invalid: any[] }
) {
    valid.map(value =>
        it(`should treat ${JSON.stringify(
            value
        )} (${typeof value}) as valid`, () => {
            const input = makeSimpleInput(type, value);
            expect(() => validateInput(input)).not.toThrowError();
        })
    );
    invalid.map(value =>
        it(`should treat ${JSON.stringify(
            value
        )} (${typeof value}) as invalid`, () => {
            const input = makeSimpleInput(type, value);
            expect(() => validateInput(input)).toThrowError();
        })
    );
}
describe('validateInput', () => {
    describe('boolean', () => {
        generateValidityTests(InputType.Boolean, validityTestCases.boolean);
    });

    describe('blob', () => {
        generateValidityTests(InputType.Blob, validityTestCases.blob);
    });

    describe('datetime', () => {
        generateValidityTests(InputType.Datetime, validityTestCases.datetime);
    });

    describe('duration', () => {
        generateValidityTests(InputType.Duration, validityTestCases.duration);
    });

    describe('float', () => {
        generateValidityTests(InputType.Float, validityTestCases.float);
    });

    describe('integer', () => {
        generateValidityTests(InputType.Integer, validityTestCases.integer);
    });

    describe('string', () => {
        generateValidityTests(InputType.String, validityTestCases.string);
    });

    it('should throw errors for missing required simple values', () => {
        const [type, input] = literalTestCases[0];
        const simpleInput = makeSimpleInput(type, input);
        simpleInput.required = true;
        delete simpleInput.value;
        expect(() => validateInput(simpleInput)).toThrowError();
    });

    it('should throw errors for missing required Blob values', () => {
        // URI is the only required, user-provided value with no default
        const simpleInput = makeSimpleInput(InputType.Blob, {
            format: 'csv',
            dimensionality: BlobDimensionality.SINGLE
        });
        simpleInput.required = true;
        expect(() => validateInput(simpleInput)).toThrowError();
    });

    it('should not throw an error for a required input with an initial value and no value', () => {
        const simpleInput = makeSimpleInput(
            InputType.String,
            primitiveLiteral({ stringValue: '' })
        );
        simpleInput.required = true;
        simpleInput.initialValue = primitiveLiteral({ stringValue: 'abcdefg' });
        delete simpleInput.value;
        expect(() => validateInput(simpleInput)).not.toThrowError();
    });
});
