import { obj } from 'test/utils';
import { formatRetryAttempt, getUniqueTaskExecutionName } from '../utils';

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
