import { TaskType } from './constants';
import { isMapTaskType, isMapTaskV1 } from './utils';

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

  it('isMapTaskV1 propely identifies mapped task events above version 1', () => {
    const eventVersionZero = 0;
    const eventVersionN = 3;
    const resourcesLengthZero = 0;
    const resourcesLengthN = 5;
    const mapTaskTypes = [TaskType.ARRAY, TaskType.ARRAY_AWS, TaskType.ARRAY_K8S];
    const nonMapTaskType = TaskType.PYTHON;

    describe('FALSE cases:', () => {
      // when no task is provided
      expect(isMapTaskV1(eventVersionN, resourcesLengthN)).toBeFalsy();
      // when regular task is provided
      expect(isMapTaskV1(eventVersionN, resourcesLengthN, nonMapTaskType)).toBeFalsy();
      // when eventVersion < 1
      expect(isMapTaskV1(eventVersionZero, resourcesLengthN, mapTaskTypes[0])).toBeFalsy();
      // when externalResources array has length of 0
      expect(isMapTaskV1(eventVersionN, resourcesLengthZero, mapTaskTypes[0])).toBeFalsy();
    });

    describe('TRUE cases:', () => {
      // when mapped task is provided AND eventVersion >= 1 AND externalResources array length > 0
      for (const task of mapTaskTypes) {
        expect(isMapTaskV1(eventVersionN, resourcesLengthN, task)).toBeTruthy();
      }
    });
  });
});
