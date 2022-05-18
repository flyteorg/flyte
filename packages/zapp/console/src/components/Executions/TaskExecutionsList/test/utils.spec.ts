import { getTaskLogName, getTaskIndex } from 'components/Executions/TaskExecutionsList/utils';
import { Event } from 'flyteidl';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { obj } from 'test/utils';
import {
  getTaskRetryAtemptsForIndex,
  formatRetryAttempt,
  getGroupedLogs,
  getUniqueTaskExecutionName,
} from '../utils';
import { getMockMapTaskLogItem, MockMapTaskExecution } from '../TaskExecutions.mocks';

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

  it('should properly group to Success and Failed', () => {
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

describe('getTaskRetryAttemptsForIndex', () => {
  it('should return 2 filtered attempts for provided index', () => {
    const index = 3;
    // '?? []' -> TS check, mock contains externalResources
    const result = getTaskRetryAtemptsForIndex(
      MockMapTaskExecution.closure.metadata?.externalResources ?? [],
      index,
    );
    expect(result).toHaveLength(2);
  });

  it('should return 1 filtered attempt for provided index', () => {
    const index = 0;
    // '?? []' -> TS check, mock contains externalResources
    const result = getTaskRetryAtemptsForIndex(
      MockMapTaskExecution.closure.metadata?.externalResources ?? [],
      index,
    );
    expect(result).toHaveLength(1);
  });

  it('should return empty array when null index provided', () => {
    const index = null;
    // '?? []' -> TS check, mock contains externalResources
    const result = getTaskRetryAtemptsForIndex(
      MockMapTaskExecution.closure.metadata?.externalResources ?? [],
      index,
    );
    expect(result).toHaveLength(0);
  });
});

describe('getTaskIndex', () => {
  it('should return index if selected log has a match in externalResources list', () => {
    const index = 3;
    const log = getMockMapTaskLogItem(TaskExecutionPhase.SUCCEEDED, true, index, 1).logs?.[0];

    // TS check
    if (log) {
      const result1 = getTaskIndex(MockMapTaskExecution, log);
      expect(result1).toStrictEqual(index);
    }
  });
});

describe('getTaskLogName', () => {
  it('should return correct names', () => {
    const taskName1 = 'task_name_1';
    const taskName2 = 'task.task_name_1';
    const taskLogName1 = 'abc';
    const taskLogName2 = 'abc-1';
    const taskLogName3 = 'abc-1-1';

    const result1 = getTaskLogName(taskName1, taskLogName1);
    expect(result1).toStrictEqual(taskName1);

    const result2 = getTaskLogName(taskName1, taskLogName2);
    expect(result2).toStrictEqual('task_name_1-1');

    const result3 = getTaskLogName(taskName1, taskLogName3);
    expect(result3).toStrictEqual('task_name_1-1-1');

    const result4 = getTaskLogName(taskName2, taskLogName1);
    expect(result4).toStrictEqual(taskName1);

    const result5 = getTaskLogName(taskName2, taskLogName2);
    expect(result5).toStrictEqual('task_name_1-1');

    const result6 = getTaskLogName(taskName2, taskLogName3);
    expect(result6).toStrictEqual('task_name_1-1-1');
  });
});
