import { Core } from 'flyteidl';
import { generateBlobType } from './literalHelpers';
import { TestCaseList } from '../types';

const blobTestcases: TestCaseList<Core.IBlob> = {
  single_CSV_BLOB: {
    value: generateBlobType('csv', Core.BlobType.BlobDimensionality.SINGLE, '1'),
    expected: {
      result_var: {
        type: 'single (csv) blob',
        uri: '1',
      },
    },
  },
  multi_part_CSV_BLOB: {
    value: generateBlobType('csv', Core.BlobType.BlobDimensionality.MULTIPART, '2'),
    expected: {
      result_var: {
        type: 'multi-part (csv) blob',
        uri: '2',
      },
    },
  },
  single_blob_BLOB: {
    value: generateBlobType(undefined, Core.BlobType.BlobDimensionality.SINGLE, '3'),
    expected: {
      result_var: {
        type: 'single blob',
        uri: '3',
      },
    },
  },
  single_multi_part_BLOB: {
    value: generateBlobType(undefined, Core.BlobType.BlobDimensionality.MULTIPART, '4'),
    expected: {
      result_var: {
        type: 'multi-part blob',
        uri: '4',
      },
    },
  },
};

export default blobTestcases;
