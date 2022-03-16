import { dateToTimestamp } from 'common/utils';
import { timeStampOffset } from 'mocks/utils';

// Arbitrary start date used as a basis for generating timestamps
// to keep the mocked data consistent across runs
export const mockStartDate = new Date('2020-11-15T02:32:19.610Z');

// entities will be created an hour before any executions have run
export const entityCreationDate = timeStampOffset(dateToTimestamp(mockStartDate), -3600);

// Workflow Execution duration in milliseconds
export const defaultExecutionDuration = 1000 * 60 * 60 * 1.251;

export const dataUriPrefix = 's3://flytedata';
export const emptyInputUri = `${dataUriPrefix}/empty/inputs.pb`;
export const emptyOutputUri = `${dataUriPrefix}/empty/outputs.pb`;

export const testProject = 'flytetest';
export const testDomain = 'development';

export const testVersions = {
  v1: 'v0001',
  v2: 'v0002',
};

export const variableNames = {
  basicString: 'basicString',
};

export const nodeIds = {
  dynamicTask: 'dynamicTask',
  pythonTask: 'pythonTask',
};
