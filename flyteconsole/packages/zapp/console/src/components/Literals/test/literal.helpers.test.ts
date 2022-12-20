import { Core } from 'flyteidl';
import { transformLiterals } from '../helpers';

import {
  primitive,
  blob,
  binary,
  schema,
  noneType,
  errorType,
  generic,
  structuredDataset,
  collection,
  map,
} from './helpers/index';
import { getCollection, getMap, getScalarLiteral } from './helpers/literalHelpers';

const literalTestCases = {
  scalar: {
    primitive,
    blob,
    binary,
    schema,
    noneType,
    error: errorType,
    generic,
    structuredDataset,
    union: {} as Core.IPrimitive, // TODO: FC#450 ass support for union types
  },
  collection,
  map,
};

describe('scalar literal', () => {
  const scalarTestCases = literalTestCases.scalar;
  const scalarType = Object.keys(scalarTestCases);

  scalarType.map((scalarTestType) => {
    describe(scalarTestType, () => {
      const cases = scalarTestCases[scalarTestType];

      Object.keys(cases || {}).map((testKey) => {
        const { value, expected } = cases[testKey];

        it(`${testKey}: should return ${expected} for ${value}`, () => {
          const scalar = { result_var: { ...getScalarLiteral(value, scalarTestType as any) } };
          const result = transformLiterals(scalar as any);
          expect(result).toEqual(expected);
        });
      });
    });
  });
});

describe('collection literal', () => {
  const cases = literalTestCases.collection;

  Object.keys(cases).map((testKey) => {
    const { value, expected } = cases[testKey];

    it(`${testKey}: should return ${expected} for ${value}`, () => {
      const collection = { result_var: { ...getCollection([value]) } };

      const result = transformLiterals(collection);
      expect(result).toEqual(expected);
    });
  });
});

describe('map literal', () => {
  const cases = literalTestCases.map;

  Object.keys(cases).map((testKey) => {
    const { value, expected } = cases[testKey];

    it(`${testKey}: should return ${expected} for ${value}`, () => {
      const collection = { result_var: { ...getMap({ value }) } };

      const result = transformLiterals(collection);
      expect(result).toEqual(expected);
    });
  });
});
