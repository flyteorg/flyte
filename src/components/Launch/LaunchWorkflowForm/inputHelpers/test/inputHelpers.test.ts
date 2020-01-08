import { InputProps, InputType } from '../../types';
import { inputToLiteral, validateInput } from '../inputHelpers';
import { literalTestCases, validityTestCases } from './testCases';

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
    // TODO: test cases for all of the input types

    // TODO
    it('should return system default if parsing literal fails', () => {});
});

describe('inputToLiteral', () => {
    describe('Primitives', () => {
        literalTestCases.map(([type, input, output]) =>
            it(`Should correctly convert ${type}: ${input} (${typeof input})`, () => {
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
            it(`Should return empty value for type: ${type}`, () => {
                expect(
                    inputToLiteral(makeSimpleInput(type, '')).scalar
                ).toEqual({ noneType: {} });
            })
        );
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
});
