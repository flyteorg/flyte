import { stringifyValue } from 'common/utils';
import { Core } from 'flyteidl';
import { InputTypeDefinition, InputValue } from '../types';
import { getHelperForInput } from './getHelperForInput';
import { ConverterInput, InputHelper, InputValidatorParams } from './types';
import t from '../../../common/strings';
import { parseJSON } from './parseJson';
import { literalNone } from './constants';

function parseMap(map: string) {
  const parsed = parseJSON(map);
  if (typeof parsed !== 'object') {
    throw new Error(t('valueNotParse'));
  }
  return parsed;
}

function fromLiteral(literal: Core.ILiteral, { subtype }: InputTypeDefinition): InputValue {
  if (!subtype) {
    throw new Error(t('missingMapSubType'));
  }
  if (!literal.map) {
    throw new Error(t('mapMissingMapProperty'));
  }
  if (!literal.map.literals) {
    throw new Error(t('mapMissingMapLiteralsProperty'));
  }
  if (typeof literal.map.literals !== 'object') {
    throw new Error(t('mapLiternalNotObject'));
  }
  if (!Object.keys(literal.map.literals).length) {
    throw new Error(t('mapLiternalObjectEmpty'));
  }

  const result = {};

  Object.entries(literal.map.literals).forEach(([key, childLiteral]) => {
    const helper = getHelperForInput(subtype.type);
    result[key] = helper.fromLiteral(childLiteral, subtype);
  });

  return stringifyValue(result);
}

function toLiteral({ value, typeDefinition: { subtype } }: ConverterInput): Core.ILiteral {
  if (!subtype) {
    throw new Error(t('missingMapSubType'));
  }
  let parsed: { [key: string]: any };
  // If we're processing a nested map, it may already have been parsed
  if (typeof value === 'object') {
    parsed = value;
  } else {
    const stringValue = typeof value === 'string' ? value : value.toString();
    if (!stringValue.length) {
      return literalNone();
    }
    parsed = parseMap(stringValue);
  }
  const result = { map: { literals: {} } };
  Object.keys(parsed)?.forEach((key) => {
    const helper = getHelperForInput(subtype.type);
    result.map.literals[key] = helper.toLiteral({ value: parsed[key], typeDefinition: subtype });
  });

  return result;
}

function validate({ value, typeDefinition: { subtype } }: InputValidatorParams) {
  if (!subtype) {
    throw new Error(t('missingMapSubType'));
  }
  if (typeof value !== 'string') {
    throw new Error(t('valueNotString'));
  }
  if (!value.toString().length) {
    throw new Error(t('valueRequired'));
  }
  try {
    parseMap(value);
  } catch (e) {
    throw new Error(t('valueNotParse'));
  }
  const obj = parseJSON(value);
  if (!Object.keys(obj).length || Object.keys(obj).some((key) => !key.trim().length)) {
    throw new Error(t('valueKeyRequired'));
  }
  Object.keys(obj).forEach((key) => {
    const helper = getHelperForInput(subtype.type);
    const subValue = obj[key];

    try {
      helper.validate({ value: subValue, typeDefinition: subtype, name: '', required: false });
    } catch (e) {
      throw new Error(t('valueValueInvalid'));
    }
  });
}

export const mapHelper: InputHelper = {
  fromLiteral,
  toLiteral,
  validate,
};
