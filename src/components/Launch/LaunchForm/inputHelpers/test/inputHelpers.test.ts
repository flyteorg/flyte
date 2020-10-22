import { Core } from 'flyteidl';
import * as Long from 'long';
import { BlobDimensionality } from 'models';
import {
    collectionInputTypeDefinition,
    nestedCollectionInputTypeDefinition,
    primitiveLiteral
} from '../../__mocks__/utils';
import { InputProps, InputType, InputTypeDefinition } from '../../types';
import { literalNone } from '../constants';
import { getHelperForInput } from '../getHelperForInput';
import {
    inputToLiteral,
    literalToInputValue,
    validateInput
} from '../inputHelpers';
import { collectionChildToString } from '../utils';
import {
    inputTypes,
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
    typeDefinition: inputTypes.unknown
};

function makeSimpleInput(
    typeDefinition: InputTypeDefinition,
    value: any
): InputProps {
    return { ...baseInputProps, value, typeDefinition };
}

function makeCollectionInput(
    typeDefinition: InputTypeDefinition,
    value: string
): InputProps {
    return {
        ...baseInputProps,
        value,
        typeDefinition: collectionInputTypeDefinition(typeDefinition)
    };
}

function makeNestedCollectionInput(
    typeDefinition: InputTypeDefinition,
    value: string
): InputProps {
    return {
        ...baseInputProps,
        value,
        typeDefinition: nestedCollectionInputTypeDefinition(typeDefinition)
    };
}

describe('literalToInputValue', () => {
    describe('Primitives', () => {
        literalToInputTestCases.map(([typeDefinition, input, output]) =>
            it(`should correctly convert ${
                typeDefinition.type
            }: ${JSON.stringify(input)}`, () =>
                expect(literalToInputValue(typeDefinition, input)).toEqual(
                    output
                ))
        );

        supportedPrimitives.map(typeDefinition =>
            it(`should convert None value for ${typeDefinition.type} to undefined`, () =>
                expect(
                    literalToInputValue(typeDefinition, literalNone())
                ).toBeUndefined())
        );

        it('should correctly convert noneType to undefined', () =>
            expect(literalToInputValue(inputTypes.none, literalNone())).toEqual(
                undefined
            ));
    });

    describe('Collections', () => {
        literalToInputTestCases.map(([typeDefinition, input, output]) => {
            it(`should correctly convert collection of ${
                typeDefinition.type
            }: ${JSON.stringify(input)}`, () => {
                const collection: Core.ILiteral = {
                    collection: {
                        // Duplicate it to test comma separation
                        literals: [input, input]
                    }
                };
                const stringifiedValue = collectionChildToString(
                    typeDefinition.type,
                    output
                );
                const expectedString = `[${stringifiedValue},${stringifiedValue}]`;
                const result = literalToInputValue(
                    collectionInputTypeDefinition(typeDefinition),
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

            const typeDefinition: InputTypeDefinition = {
                literalType: { simple: Core.SimpleType.NONE },
                type: InputType.None
            };

            expect(
                literalToInputValue(
                    collectionInputTypeDefinition(typeDefinition),
                    collection
                )
            ).toEqual('[]');
        });
    });

    it('should return system default if parsing literal fails', () => {
        const { defaultValue } = getHelperForInput(InputType.Boolean);
        expect(
            literalToInputValue(
                inputTypes.boolean,
                // Invalid boolean input value because it uses the string field
                { scalar: { primitive: { stringValue: 'whoops' } } }
            )
        ).toEqual(defaultValue);
    });
});

describe('inputToLiteral', () => {
    describe('Scalars', () => {
        literalTestCases.map(([typeDefinition, input, output]) => {
            it(`should correctly convert ${
                typeDefinition.type
            }: ${JSON.stringify(input)} (${typeof input})`, () =>
                expect(
                    inputToLiteral(makeSimpleInput(typeDefinition, input))
                ).toEqual(output));
        });
    });

    describe('Collections', () => {
        literalTestCases.map(([typeDefinition, input, output]) => {
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

            it(`should correctly convert collection of type ${
                typeDefinition.type
            }: [${JSON.stringify(value)}] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeCollectionInput(typeDefinition, `[${value}]`)
                );
                expect(result.collection!.literals![0]).toEqual(output);
            });

            it(`should correctly convert nested collection of type ${
                typeDefinition.type
            }: [[${JSON.stringify(value)}]] (${typeof input})`, () => {
                const result = inputToLiteral(
                    makeNestedCollectionInput(typeDefinition, `[[${value}]]`)
                );
                expect(
                    result.collection!.literals![0].collection!.literals![0]
                ).toEqual(output);
            });
        });
    });

    describe('Unsupported Types', () => {
        unsupportedTypes.map(typeDefinition =>
            it(`should return empty value for type: ${typeDefinition.type}`, () => {
                expect(
                    inputToLiteral(makeSimpleInput(typeDefinition, '')).scalar
                ).toEqual({ noneType: {} });
            })
        );
    });

    it('Should return initial value for inputs with no value', () => {
        const simpleInput = makeSimpleInput(
            inputTypes.string,
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
    typeDefinition: InputTypeDefinition,
    { valid, invalid }: { valid: any[]; invalid: any[] }
) {
    valid.map(value =>
        it(`should treat ${JSON.stringify(
            value
        )} (${typeof value}) as valid`, () => {
            const input = makeSimpleInput(typeDefinition, value);
            expect(() => validateInput(input)).not.toThrowError();
        })
    );
    invalid.map(value =>
        it(`should treat ${JSON.stringify(
            value
        )} (${typeof value}) as invalid`, () => {
            const input = makeSimpleInput(typeDefinition, value);
            expect(() => validateInput(input)).toThrowError();
        })
    );
}
describe('validateInput', () => {
    describe('boolean', () => {
        generateValidityTests(inputTypes.boolean, validityTestCases.boolean);
    });

    describe('blob', () => {
        generateValidityTests(inputTypes.blobSingle, validityTestCases.blob);
    });

    describe('datetime', () => {
        generateValidityTests(inputTypes.datetime, validityTestCases.datetime);
    });

    describe('duration', () => {
        generateValidityTests(inputTypes.duration, validityTestCases.duration);
    });

    describe('float', () => {
        generateValidityTests(inputTypes.float, validityTestCases.float);
    });

    describe('integer', () => {
        generateValidityTests(inputTypes.integer, validityTestCases.integer);
    });

    describe('schema', () => {
        generateValidityTests(inputTypes.schema, validityTestCases.schema);
    });

    describe('string', () => {
        generateValidityTests(inputTypes.string, validityTestCases.string);
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
        const simpleInput = makeSimpleInput(inputTypes.blobSingle, {
            format: 'csv',
            dimensionality: BlobDimensionality.SINGLE
        });
        simpleInput.required = true;
        expect(() => validateInput(simpleInput)).toThrowError();
    });

    it('should not throw an error for a required input with an initial value and no value', () => {
        const simpleInput = makeSimpleInput(
            inputTypes.string,
            primitiveLiteral({ stringValue: '' })
        );
        simpleInput.required = true;
        simpleInput.initialValue = primitiveLiteral({ stringValue: 'abcdefg' });
        delete simpleInput.value;
        expect(() => validateInput(simpleInput)).not.toThrowError();
    });
});
