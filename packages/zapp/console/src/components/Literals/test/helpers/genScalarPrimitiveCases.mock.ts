import { Core } from 'flyteidl';
import * as Long from 'long';
import { long } from 'test/utils';
import { getPrimitive } from './literalHelpers';
import { TestCaseList } from '../types';

const scalarPrimitiveTestCases: TestCaseList<Core.IPrimitive> = {
  INT_WITH_LARGE_LOW: {
    value: getPrimitive('integer', { low: 1642627611, high: 0, unsigned: false } as Long),
    expected: { result_var: 1642627611 },
  },
  INT_WITH_SMALL_LOW: {
    value: getPrimitive('integer', { low: 55, high: 0, unsigned: false } as Long),
    expected: { result_var: 55 },
  },
  INT_WITH_ZERO: {
    value: getPrimitive('integer', { low: 0, high: 0, unsigned: false } as Long),
    expected: { result_var: 0 },
  },
  FLOAT_ZERO: {
    value: getPrimitive('floatValue', 0),
    expected: { result_var: 0 },
  },
  FLOAT_POSITIVE: {
    value: getPrimitive('floatValue', 1.5),
    expected: { result_var: 1.5 },
  },
  FLOAT_NEGATIVE: {
    value: getPrimitive('floatValue', -1.5),
    expected: { result_var: -1.5 },
  },
  FLOAT_LARGE: {
    value: getPrimitive('floatValue', 1.25e10),
    expected: { result_var: 1.25e10 },
  },
  FLOAT_UNDEF: {
    value: getPrimitive('floatValue', undefined),
    expected: { result_var: undefined },
  },
  STRING_NULL: {
    value: getPrimitive('stringValue', null),
    expected: { result_var: 'null' },
  },
  STRING_VALID: {
    value: getPrimitive('stringValue', 'some val'),
    expected: { result_var: 'some val' },
  },
  STRING_UNDEF: {
    value: getPrimitive('stringValue', undefined),
    expected: { result_var: 'undefined' },
  },
  STRING_NEG: {
    value: getPrimitive('stringValue', '-1'),
    expected: { result_var: '-1' },
  },
  BOOL_TRUE: {
    value: getPrimitive('boolean', true),
    expected: { result_var: true },
  },
  BOOL_FALSE: {
    value: getPrimitive('boolean', false),
    expected: { result_var: false },
  },
  BOOL_NULL: {
    value: getPrimitive('boolean', null),
    expected: { result_var: null },
  },
  BOOL_UNDEF: {
    value: getPrimitive('boolean', undefined),
    expected: { result_var: undefined },
  },
  DT_VALID: {
    value: getPrimitive('datetime', { seconds: long(3600), nanos: 0 }),
    expected: { result_var: '1/1/1970 1:00:00 AM UTC' },
  },
  DURATION_ZERO: {
    value: getPrimitive('duration', { seconds: long(0), nanos: 0 }),
    expected: { result_var: '0s' },
  },
  DURATION_1H: {
    value: getPrimitive('duration', { seconds: long(3600), nanos: 0 }),
    expected: { result_var: '1h' },
  },
  DURATION_LARGE: {
    value: getPrimitive('duration', { seconds: long(10000), nanos: 0 }),
    expected: { result_var: '2h 46m 40s' },
  },
};

export default scalarPrimitiveTestCases;
