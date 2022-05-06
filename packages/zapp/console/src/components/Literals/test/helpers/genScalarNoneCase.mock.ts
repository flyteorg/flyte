import { Core } from 'flyteidl';
import { TestCaseList } from '../types';

const scalarNoneTestCase: TestCaseList<Core.IVoid> = {
  VOID_TYPE: {
    value: {},
    expected: {
      result_var: '(empty)',
    },
  },
};

export default scalarNoneTestCase;
