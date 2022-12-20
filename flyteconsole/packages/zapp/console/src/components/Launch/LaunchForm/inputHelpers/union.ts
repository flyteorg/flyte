import { Core } from 'flyteidl';
import { isObject } from 'lodash';
import { InputTypeDefinition, InputValue, UnionValue } from '../types';
import { getHelperForInput } from './getHelperForInput';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import t from '../../../common/strings';

function fromLiteral(literal: Core.ILiteral, inputTypeDefinition: InputTypeDefinition): InputValue {
  const { listOfSubTypes } = inputTypeDefinition;
  if (!listOfSubTypes?.length) {
    throw new Error(t('missingUnionListOfSubType'));
  }

  // loop though the subtypes to find the correct match literal type
  for (let i = 0; i < listOfSubTypes.length; i++) {
    try {
      const value = getHelperForInput(listOfSubTypes[i].type).fromLiteral(
        literal,
        listOfSubTypes[i],
      );
      return { value, typeDefinition: listOfSubTypes[i] } as UnionValue;
    } catch (error) {
      // do nothing here. it's expected to have error from fromLiteral
      // because we loop through all the type to decode the input value
      // the error should be something like this
      // new Error(`Failed to extract literal value with path ${path}`);
    }
  }
  throw new Error(t('noMatchingResults'));
}

function toLiteral({ value, typeDefinition: { listOfSubTypes } }: ConverterInput): Core.ILiteral {
  if (!listOfSubTypes) {
    throw new Error(t('missingUnionListOfSubType'));
  }

  if (!isObject(value)) {
    throw new Error(t('valueMustBeObject'));
  }

  const { value: unionValue, typeDefinition } = value as UnionValue;

  return getHelperForInput(typeDefinition.type).toLiteral({
    value: unionValue,
    typeDefinition: typeDefinition,
  } as ConverterInput);
}

function validate({ value, typeDefinition: { listOfSubTypes } }: InputValidatorParams) {
  if (!value) {
    throw new Error(t('valueRequired'));
  }
  if (!isObject(value)) {
    throw new Error(t('valueMustBeObject'));
  }

  const { typeDefinition } = value as UnionValue;
  getHelperForInput(typeDefinition.type).validate(value as InputValidatorParams);
}

export const unionHelper: InputHelper = {
  fromLiteral,
  toLiteral,
  validate,
};
