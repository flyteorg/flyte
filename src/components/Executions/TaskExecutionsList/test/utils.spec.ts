import { obj } from 'test/utils';
import { getUniqueTaskExecutionName } from '../utils';

describe('getUniqueTaskExecutionName', () => {
    const cases: [{ name: string; retryAttempt: number }, string][] = [
        [{ name: 'task1', retryAttempt: 0 }, 'task1'],
        [{ name: 'task1', retryAttempt: 1 }, 'task1 (2)'],
        [{ name: 'task1', retryAttempt: 2 }, 'task1 (3)']
    ];

    cases.forEach(([input, expected]) =>
        it(`should return ${expected} for ${obj(input)}`, () =>
            expect(
                getUniqueTaskExecutionName({
                    id: {
                        retryAttempt: input.retryAttempt,
                        taskId: { name: input.name }
                    }
                } as any)
            ).toEqual(expected))
    );
});
