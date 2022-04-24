import {
  collectionInputTypeDefinition,
  nestedCollectionInputTypeDefinition,
} from '../../__mocks__/utils';
import { InputTypeDefinition } from '../../types';
import { typeIsSupported } from '../utils';
import { supportedPrimitives, unsupportedTypes } from './testCases';

type TypeIsSupportedTestCase = [string, InputTypeDefinition, boolean];
describe('Launch/inputHelpers/utils', () => {
  describe('typeIsSupported', () => {
    const cases: TypeIsSupportedTestCase[] = [
      ...supportedPrimitives.map<TypeIsSupportedTestCase>((typeDefinition) => [
        `supports type ${typeDefinition.type}`,
        typeDefinition,
        true,
      ]),
      ...supportedPrimitives.map<TypeIsSupportedTestCase>((typeDefinition) => [
        `supports 1-dimension collection of type ${typeDefinition.type}`,
        collectionInputTypeDefinition(typeDefinition),
        true,
      ]),
      ...supportedPrimitives.map<TypeIsSupportedTestCase>((typeDefinition) => [
        `supports 2-dimension collection of type: ${typeDefinition.type}`,
        nestedCollectionInputTypeDefinition(typeDefinition),
        true,
      ]),
      ...unsupportedTypes.map<TypeIsSupportedTestCase>((typeDefinition) => [
        `does NOT support type ${typeDefinition.type}`,
        typeDefinition,
        false,
      ]),
      ...unsupportedTypes.map<TypeIsSupportedTestCase>((typeDefinition) => [
        `does NOT support 1-dimension collection of type ${typeDefinition.type}`,
        collectionInputTypeDefinition(typeDefinition),
        false,
      ]),
      ...unsupportedTypes.map<TypeIsSupportedTestCase>((typeDefinition) => [
        `does NOT support 2-dimension collection of type: ${typeDefinition.type}`,
        nestedCollectionInputTypeDefinition(typeDefinition),
        false,
      ]),
    ];

    cases.forEach(([description, value, expected]) =>
      it(description, () => expect(typeIsSupported(value)).toBe(expected)),
    );
  });
});
