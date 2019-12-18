import { Core } from 'flyteidl';
import { InputTypeDefinition, InputValue } from '../types';

export interface ConverterInput {
    value: InputValue;
    typeDefinition: InputTypeDefinition;
}

export type LiteralConverterFn = (input: ConverterInput) => Core.ILiteral;
export interface InputHelper {
    toLiteral: LiteralConverterFn;
    /** Will throw in the case of a failed validation */
    validate: (input: ConverterInput) => void;
}
