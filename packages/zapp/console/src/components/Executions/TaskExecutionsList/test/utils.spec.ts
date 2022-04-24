import { Event } from 'flyteidl';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { obj } from 'test/utils';
import { getMockMapTaskLogItem } from '../TaskExecutions.mocks';
import { formatRetryAttempt, getGroupedLogs, getUniqueTaskExecutionName } from '../utils';

describe('getUniqueTaskExecutionName', () => {
  const cases: [{ name: string; retryAttempt: number }, string][] = [
    [{ name: 'task1', retryAttempt: 0 }, 'task1'],
    [{ name: 'task1', retryAttempt: 1 }, 'task1 (2)'],
    [{ name: 'task1', retryAttempt: 2 }, 'task1 (3)'],
  ];

  cases.forEach(([input, expected]) =>
    it(`should return ${expected} for ${obj(input)}`, () =>
      expect(
        getUniqueTaskExecutionName({
          id: {
            retryAttempt: input.retryAttempt,
            taskId: { name: input.name },
          },
        } as any),
      ).toEqual(expected)),
  );
});

describe('formatRetryAttempt', () => {
  const cases: [number, string][] = [
    [0, 'Attempt 01'],
    [1, 'Attempt 02'],
    [2, 'Attempt 03'],
  ];

  cases.forEach(([input, expected]) =>
    it(`should return ${expected} for ${input}`, () =>
      expect(formatRetryAttempt(input)).toEqual(expected)),
  );
});

describe('getGroupedLogs', () => {
  const resources: Event.IExternalResourceInfo[] = [
    getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true),
    getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 1),
    getMockMapTaskLogItem(TaskExecutionPhase.FAILED, true, 1, 1),
    getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, 1, 2),
    getMockMapTaskLogItem(TaskExecutionPhase.FAILED, false, 2),
  ];

  it(`Should properly group to Success and Failed`, () => {
    const logs = getGroupedLogs(resources);
    // Do not have key which was not in the logs
    expect(logs.get(TaskExecutionPhase.QUEUED)).toBeUndefined();

    // To have keys which were in the logs
    expect(logs.get(TaskExecutionPhase.SUCCEEDED)).not.toBeUndefined();
    expect(logs.get(TaskExecutionPhase.FAILED)).not.toBeUndefined();

    // to include all items with last retry iterations
    expect(logs.get(TaskExecutionPhase.SUCCEEDED)?.length).toEqual(2);

    // to filter our previous retry attempt
    expect(logs.get(TaskExecutionPhase.FAILED)?.length).toEqual(1);
  });
});
