import { TaskType } from './constants';
import { isMapTaskType } from './utils';

describe('models/Task', () => {
  it('isMapTaskType propely identifies mapped tasks', async () => {
    // when no task is provided - to be false
    expect(isMapTaskType()).toBeFalsy();

    // when mapped task is provided - to be true
    expect(isMapTaskType(TaskType.ARRAY)).toBeTruthy();
    expect(isMapTaskType(TaskType.ARRAY_AWS)).toBeTruthy();
    expect(isMapTaskType(TaskType.ARRAY_K8S)).toBeTruthy();

    // when regular task is provided - to be false
    expect(isMapTaskType(TaskType.PYTHON)).toBeFalsy();
  });
});
