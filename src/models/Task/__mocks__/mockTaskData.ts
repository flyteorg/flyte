import { TaskType } from '../constants';
import { CompiledTask } from '../types';

export const mockTasks: CompiledTask[] = [
  {
    template: {
      type: TaskType.PYTHON,
      id: {
        project: 'flytekit',
        domain: 'development',
        name: 'BasicNode',
        version: 'abcdef123456',
      },
    },
  },
  {
    template: {
      type: TaskType.PYTHON,
      id: {
        project: 'flytekit',
        domain: 'development',
        name: 'PythonNode',
        version: 'abcdef123456',
      },
    },
  },
  {
    template: {
      type: TaskType.HIVE,
      id: {
        project: 'flytekit',
        domain: 'development',
        name: 'HiveNode',
        version: 'abcdef123456',
      },
    },
  },
  {
    template: {
      type: TaskType.DYNAMIC,
      id: {
        project: 'flytekit',
        domain: 'development',
        name: 'UnexpectedlyVerboseAndLengthyNameNode',
        version: 'abcdef123456',
      },
    },
  },
  {
    template: {
      type: TaskType.DYNAMIC,
      id: {
        project: 'flytekit',
        domain: 'development',
        name: 'DynamicTaskNode',
        version: 'abcdef123456',
      },
    },
  },
];
