import { Core } from 'flyteidl';
import { InputValue } from '../types';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';

function fromLiteral(literal: Core.ILiteral): InputValue {
  return {};
}

function toLiteral({ value }: ConverterInput): Core.ILiteral {
  return {
    scalar: { noneType: {} },
  };
}

function validate({ value }: InputValidatorParams) {}

export const noneHelper: InputHelper = {
  fromLiteral,
  toLiteral,
  validate,
};
