import { Blob, BlobDimensionality } from 'models/Common/types';

export const blobValues: Dictionary<Blob> = {
  singleBlobValue: {
    metadata: {
      type: { dimensionality: BlobDimensionality.SINGLE },
    },
    uri: 's3://path/to/my/single_blob',
  },
  singleBlobValueWithCSVFormat: {
    metadata: {
      type: { dimensionality: BlobDimensionality.SINGLE, format: 'CSV' },
    },
    uri: 's3://path/to/my/single_csv_blob',
  },
  multiPartBlobValue: {
    metadata: {
      type: { dimensionality: BlobDimensionality.MULTIPART },
    },
    uri: 's3://path/to/my/multi_blob_prefix',
  },
  multiPartBlobValueWithCSVFormat: {
    metadata: {
      type: {
        dimensionality: BlobDimensionality.MULTIPART,
        format: 'CSV',
      },
    },
    uri: 's3://path/to/my/multi_csv_blob_prefix',
  },
};
