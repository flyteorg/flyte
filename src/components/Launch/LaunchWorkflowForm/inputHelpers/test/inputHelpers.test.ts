import { Core } from 'flyteidl';
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
                input.scalar!.primitive
            )}`, () => {
                const result = literalToInputValue({ type }, input);
                expect(result).toEqual(output);
            })
        );

        [
            InputType.Collection,
            InputType.Datetime,
            InputType.Duration,
            InputType.Float,
            InputType.Integer,
            InputType.String
        ].map(type =>
            it(`should convert None value for ${type} to undefined`, () => {
                expect(
                    literalToInputValue({ type }, literalNone())
                ).toBeUndefined();
            })
        );

        it('should correctly convert noneType to undefined', () => {
            expect(
                literalToInputValue({ type: InputType.None }, literalNone())
            ).toEqual(undefined);
        });
    });

    describe('Collections', () => {
        literalToInputTestCases.map(([type, input, output]) =>
            it(`should correctly convert collection of ${type}: ${JSON.stringify(
                input.scalar!.primitive
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
            })
        );

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
    describe('Primitives', () => {
        literalTestCases.map(([type, input, output]) =>
            it(`should correctly convert ${type}: ${input} (${typeof input})`, () => {
                const result = inputToLiteral(makeSimpleInput(type, input));
                expect(result.scalar!.primitive).toEqual(output);
            })
        );
    });

    describe('Collections', () => {
        literalTestCases.map(([type, input, output]) => {
            let value: any;
            if (['boolean', 'number'].includes(typeof input)) {
                value = input;
            } else if (input instanceof Date) {
                value = `"${input.toISOString()}"`;
            } else {
                value = `"${input}"`;
            }

            it(`should correctly convert collection of type ${type}: [${value}] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeCollectionInput(type, `[${value}]`)
                );
                expect(
                    result.collection!.literals![0].scalar!.primitive
                ).toEqual(output);
            });

            it(`should correctly convert nested collection of type ${type}: [[${value}]] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeNestedCollectionInput(type, `[[${value}]]`)
                );
                expect(
                    result.collection!.literals![0].collection!.literals![0]
                        .scalar!.primitive
                ).toEqual(output);
            });
        });
    });

    describe('Unsupported Types', () => {
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
        it(`should treat ${value} (${typeof value}) as valid`, () => {
            const input = makeSimpleInput(type, value);
            expect(() => validateInput(input)).not.toThrowError();
        })
    );
    invalid.map(value =>
        it(`should treat ${value} (${typeof value}) as invalid`, () => {
            const input = makeSimpleInput(type, value);
            expect(() => validateInput(input)).toThrowError();
        })
    );
}
describe('validateInput', () => {
    describe('boolean', () => {
        generateValidityTests(InputType.Boolean, validityTestCases.boolean);
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

    it('should throw errors for missing required values', () => {
        const [type, input] = literalTestCases[0];
        const simpleInput = makeSimpleInput(type, input);
        simpleInput.required = true;
        delete simpleInput.value;
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
