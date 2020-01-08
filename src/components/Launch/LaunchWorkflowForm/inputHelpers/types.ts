import { Core } from 'flyteidl';
import { Literal } from 'models';
import { InputTypeDefinition, InputValue } from '../types';

export interface ConverterInput {
    value: InputValue;
    typeDefinition: InputTypeDefinition;
}

export type InputToLiteralConverterFn = (
    input: ConverterInput
) => Core.ILiteral;
export type LiteralToInputConterterFn = (literal: Literal) => InputValue;
export interface InputHelper {
    defaultValue?: InputValue;
    toLiteral: InputToLiteralConverterFn;
    fromLiteral: LiteralToInputConterterFn;
    /** Will throw in the case of a failed validation */
    validate: (input: ConverterInput) => void;
}
