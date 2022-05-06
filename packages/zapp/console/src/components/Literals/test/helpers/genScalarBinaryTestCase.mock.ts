import { Core } from 'flyteidl';
import { TestCaseList } from '../types';

const scalarBinaryTestCases: TestCaseList<Core.IBinary> = {
  WITH_VAL: {
    value: { value: new Uint8Array(), tag: 'tag1' },
    expected: { result_var: { tag: 'tag1 (binary data not shown)' } },
  },
  INT_WITH_SMALL_LOW: {
    value: { tag: 'tag2' },
    expected: { result_var: { tag: 'tag2 (binary data not shown)' } },
  },
};

export default scalarBinaryTestCases;
