import { Core } from 'flyteidl';
import { InputProps, InputTypeDefinition, InputValue } from '../types';

export interface ConverterInput {
  value: InputValue;
  typeDefinition: InputTypeDefinition;
}

export type InputToLiteralConverterFn = (input: ConverterInput) => Core.ILiteral;
export type LiteralToInputConterterFn = (
  literal: Core.ILiteral,
  typeDefinition: InputTypeDefinition,
) => InputValue | undefined;

export type ValidationParams = Pick<
  InputProps,
  'initialValue' | 'name' | 'required' | 'typeDefinition' | 'value'
>;

export type InputValidatorParams = ValidationParams & { value: InputValue };

export interface InputHelper {
  defaultValue?: InputValue;
  toLiteral: InputToLiteralConverterFn;
  fromLiteral: LiteralToInputConterterFn;
  /** Will throw in the case of a failed validation */
  validate: (params: InputValidatorParams) => void;
}
