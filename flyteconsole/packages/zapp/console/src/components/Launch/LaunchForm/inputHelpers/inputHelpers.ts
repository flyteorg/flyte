import { createDebugLogger } from 'common/log';
import { ValueError } from 'errors/parameterErrors';
import { Core } from 'flyteidl';
import { InputProps, InputTypeDefinition, InputValue } from '../types';
import { literalNone } from './constants';
import { getHelperForInput } from './getHelperForInput';
import { InputValidatorParams, ValidationParams } from './types';

const debug = createDebugLogger('@InputHelper');

type ToLiteralParams = Pick<InputProps, 'initialValue' | 'typeDefinition' | 'value'>;
/** Converts a type/InputValue combination to a `Core.ILiteral` which can be
 * submitted to Admin for creating an execution.
 */
export function inputToLiteral({
  initialValue,
  typeDefinition,
  value,
}: ToLiteralParams): Core.ILiteral {
  if (value == null) {
    return initialValue != null ? initialValue : literalNone();
  }

  const { toLiteral } = getHelperForInput(typeDefinition.type);
  return toLiteral({ value, typeDefinition });
}

/** Generates the default value (if any) for a given type. */
export function defaultValueForInputType(
  typeDefinition: InputTypeDefinition,
): InputValue | undefined {
  return getHelperForInput(typeDefinition.type).defaultValue;
}

/** Converts a Core.ILiteral to an InputValue which can be used to populate
 * a form control.
 */
export function literalToInputValue(
  typeDefinition: InputTypeDefinition,
  literal: Core.ILiteral,
): InputValue | undefined {
  const { defaultValue, fromLiteral } = getHelperForInput(typeDefinition.type);

  if (literal.scalar && literal.scalar.noneType) {
    return undefined;
  }

  try {
    return fromLiteral(literal, typeDefinition);
  } catch (e) {
    // If something goes wrong (most likely malformed default value input),
    // we'll return the system default value.
    debug((e as Error).message);
    return defaultValue;
  }
}

/** Validates a given InputValue based on rules for the provided type. Returns
 * void if no errors, throws an error otherwise.
 */
export function validateInput(params: ValidationParams) {
  const { initialValue, name, required, typeDefinition, value } = params;
  const { validate } = getHelperForInput(typeDefinition.type);
  if (value == null) {
    if (required && initialValue == null) {
      throw new ValueError(name, 'Value is required');
    }
    return;
  }

  try {
    validate(params as InputValidatorParams);
  } catch (e) {
    const error = e as Error;
    throw new ValueError(name, error.message);
  }
}
