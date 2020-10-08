import { InputType, InputTypeDefinition } from '../../types';
import { typeIsSupported } from '../utils';
import { supportedPrimitives, unsupportedTypes } from './testCases';

type TypeIsSupportedTestCase = [string, InputTypeDefinition, boolean];
describe('Launch/inputHelpers/utils', () => {
    describe('typeIsSupported', () => {
        const cases: TypeIsSupportedTestCase[] = [
            ...supportedPrimitives.map<TypeIsSupportedTestCase>(type => [
                `supports type ${type}`,
                { type },
                true
            ]),
            ...supportedPrimitives.map<TypeIsSupportedTestCase>(type => [
                `supports 1-dimension collection of type ${type}`,
                { type: InputType.Collection, subtype: { type } },
                true
            ]),
            ...supportedPrimitives.map<TypeIsSupportedTestCase>(type => [
                `supports 2-dimension collection of type: ${type}`,
                {
                    type: InputType.Collection,
                    subtype: { type: InputType.Collection, subtype: { type } }
                },
                true
            ]),
            ...unsupportedTypes.map<TypeIsSupportedTestCase>(type => [
                `does NOT support type ${type}`,
                { type },
                false
            ]),
            ...unsupportedTypes.map<TypeIsSupportedTestCase>(type => [
                `does NOT support 1-dimension collection of type ${type}`,
                { type: InputType.Collection, subtype: { type } },
                false
            ]),
            ...unsupportedTypes.map<TypeIsSupportedTestCase>(type => [
                `does NOT support 2-dimension collection of type: ${type}`,
                {
                    type: InputType.Collection,
                    subtype: { type: InputType.Collection, subtype: { type } }
                },
                false
            ])
        ];

        cases.forEach(([description, value, expected]) =>
            it(description, () => expect(typeIsSupported(value)).toBe(expected))
        );
    });
});
