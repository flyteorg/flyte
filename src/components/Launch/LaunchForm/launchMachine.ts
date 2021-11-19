import { Admin, Core } from 'flyteidl';
import { Identifier, NamedEntityIdentifier } from 'models/Common/types';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { LaunchPlan } from 'models/Launch/types';
import { Task } from 'models/Task/types';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import {
    assign,
    DoneInvokeEvent,
    Machine,
    MachineConfig,
    MachineOptions,
    StatesConfig
} from 'xstate';
import { LiteralValueMap, ParsedInput } from './types';

export type SelectWorkflowVersionEvent = {
    type: 'SELECT_WORKFLOW_VERSION';
    workflowId: WorkflowId;
};
export type SelectTaskVersionEvent = {
    type: 'SELECT_TASK_VERSION';
    taskId: Identifier;
};
export type SelectLaunchPlanEvent = {
    type: 'SELECT_LAUNCH_PLAN';
    launchPlan: LaunchPlan;
};
export type WorkflowVersionOptionsLoadedEvent = DoneInvokeEvent<Workflow[]>;
export type LaunchPlanOptionsLoadedEvent = DoneInvokeEvent<LaunchPlan[]>;
export type TaskVersionOptionsLoadedEvent = DoneInvokeEvent<Task[]>;
export type ExecutionCreatedEvent = DoneInvokeEvent<
    WorkflowExecutionIdentifier
>;
export type InputsParsedEvent = DoneInvokeEvent<{
    parsedInputs: ParsedInput[];
    unsupportedRequiredInputs?: ParsedInput[];
}>;
export type ErrorEvent = DoneInvokeEvent<Error>;

export type BaseLaunchEvent =
    | { type: 'CANCEL' }
    | { type: 'SUBMIT' }
    | { type: 'RETRY' }
    | InputsParsedEvent
    | SelectTaskVersionEvent
    | SelectLaunchPlanEvent
    | ExecutionCreatedEvent
    | ErrorEvent;

export type TaskLaunchEvent =
    | BaseLaunchEvent
    | TaskVersionOptionsLoadedEvent
    | SelectTaskVersionEvent;

export type WorkflowLaunchEvent =
    | BaseLaunchEvent
    | SelectWorkflowVersionEvent
    | WorkflowVersionOptionsLoadedEvent
    | LaunchPlanOptionsLoadedEvent;

export interface BaseLaunchContext {
    defaultInputValues?: LiteralValueMap;
    parsedInputs: ParsedInput[];
    resultExecutionId?: WorkflowExecutionIdentifier;
    sourceId?: NamedEntityIdentifier;
    error?: Error;
    showErrors: boolean;
    referenceExecutionId?: WorkflowExecutionIdentifier;
    unsupportedRequiredInputs: ParsedInput[];
}

export interface WorkflowLaunchContext extends BaseLaunchContext {
    launchPlan?: LaunchPlan;
    launchPlanOptions?: LaunchPlan[];
    preferredLaunchPlanId?: Identifier;
    preferredWorkflowId?: Identifier;
    workflowVersion?: WorkflowId;
    workflowVersionOptions?: Workflow[];
    defaultAuthRole?: Admin.IAuthRole;
    disableAll?: boolean | null;
    maxParallelism?: number | null;
    labels?: Admin.ILabels | null;
    annotations?: Admin.IAnnotations | null;
}

export interface TaskLaunchContext extends BaseLaunchContext {
    defaultAuthRole?: Admin.IAuthRole;
    preferredTaskId?: Identifier;
    taskVersion?: Identifier;
    taskVersionOptions?: Task[];
}

export enum LaunchState {
    CANCELLED = 'CANCELLED',
    SELECT_SOURCE = 'SELECT_SOURCE',
    LOADING_WORKFLOW_VERSIONS = 'LOADING_WORKFLOW_VERSIONS',
    FAILED_LOADING_WORKFLOW_VERSIONS = 'FAILED_LOADING_WORKFLOW_VERSIONS',
    SELECT_WORKFLOW_VERSION = 'SELECT_WORKFLOW_VERSION',
    LOADING_LAUNCH_PLANS = 'LOADING_LAUNCH_PLANS',
    FAILED_LOADING_LAUNCH_PLANS = 'FAILED_LOADING_LAUNCH_PLANS',
    SELECT_LAUNCH_PLAN = 'SELECT_LAUNCH_PLAN',
    LOADING_TASK_VERSIONS = 'LOADING_TASK_VERSIONS',
    FAILED_LOADING_TASK_VERSIONS = 'FAILED_LOADING_TASK_VERSIONS',
    SELECT_TASK_VERSION = 'SELECT_TASK_VERSION',
    LOADING_INPUTS = 'LOADING_INPUTS',
    FAILED_LOADING_INPUTS = 'FAILED_LOADING_INPUTS',
    UNSUPPORTED_INPUTS = 'UNSUPPORTED_INPUTS',
    ENTER_INPUTS = 'ENTER_INPUTS',
    VALIDATING_INPUTS = 'VALIDATING_INPUTS',
    INVALID_INPUTS = 'INVALID_INPUTS',
    SUBMIT_VALIDATING = 'SUBMIT_VALIDATING',
    SUBMITTING = 'SUBMITTING',
    SUBMIT_FAILED = 'SUBMIT_FAILED',
    SUBMIT_SUCCEEDED = 'SUBMIT_SUCCEEDED'
}

interface BaseLaunchSchema {
    states: {
        [LaunchState.CANCELLED]: {};
        [LaunchState.LOADING_INPUTS]: {};
        [LaunchState.FAILED_LOADING_INPUTS]: {};
        [LaunchState.UNSUPPORTED_INPUTS]: {};
        [LaunchState.ENTER_INPUTS]: {};
        [LaunchState.VALIDATING_INPUTS]: {};
        [LaunchState.INVALID_INPUTS]: {};
        [LaunchState.SUBMIT_VALIDATING]: {};
        [LaunchState.SUBMITTING]: {};
        [LaunchState.SUBMIT_FAILED]: {};
        [LaunchState.SUBMIT_SUCCEEDED]: {};
    };
}

interface TaskLaunchSchema extends BaseLaunchSchema {
    states: BaseLaunchSchema['states'] & {
        [LaunchState.LOADING_TASK_VERSIONS]: {};
        [LaunchState.FAILED_LOADING_TASK_VERSIONS]: {};
        [LaunchState.SELECT_TASK_VERSION]: {};
    };
}
interface WorkflowLaunchSchema extends BaseLaunchSchema {
    states: BaseLaunchSchema['states'] & {
        [LaunchState.LOADING_WORKFLOW_VERSIONS]: {};
        [LaunchState.FAILED_LOADING_WORKFLOW_VERSIONS]: {};
        [LaunchState.SELECT_WORKFLOW_VERSION]: {};
        [LaunchState.LOADING_LAUNCH_PLANS]: {};
        [LaunchState.FAILED_LOADING_LAUNCH_PLANS]: {};
        [LaunchState.SELECT_LAUNCH_PLAN]: {};
    };
}

/** Typestates to narrow down the `context` values based on the result of
 * a `state.matches` check.
 */
export type BaseLaunchTypestate =
    | {
          value: LaunchState;
          context: BaseLaunchContext;
      }
    | {
          value: LaunchState.UNSUPPORTED_INPUTS;
          context: BaseLaunchContext & {
              parsedInputs: [];
              unsupportedRequiredInputs: [];
          };
      }
    | {
          value:
              | LaunchState.ENTER_INPUTS
              | LaunchState.VALIDATING_INPUTS
              | LaunchState.INVALID_INPUTS
              | LaunchState.SUBMIT_VALIDATING
              | LaunchState.SUBMITTING
              | LaunchState.SUBMIT_SUCCEEDED;
          context: BaseLaunchContext & {
              parsedInputs: [];
          };
      }
    | {
          value: LaunchState.SUBMIT_SUCCEEDED;
          context: BaseLaunchContext & {
              resultExecutionId: WorkflowExecutionIdentifier;
          };
      }
    | {
          value: LaunchState.SUBMIT_FAILED;
          context: BaseLaunchContext & {
              parsedInputs: ParsedInput[];
              error: Error;
          };
      }
    | {
          value:
              | LaunchState.FAILED_LOADING_INPUTS
              | LaunchState.FAILED_LOADING_LAUNCH_PLANS
              | LaunchState.FAILED_LOADING_TASK_VERSIONS
              | LaunchState.FAILED_LOADING_WORKFLOW_VERSIONS;
          context: BaseLaunchContext & {
              error: Error;
          };
      };

export type WorkflowLaunchTypestate =
    | BaseLaunchTypestate
    | {
          value: LaunchState.SELECT_WORKFLOW_VERSION;
          context: WorkflowLaunchContext & {
              sourceId: NamedEntityIdentifier;
              workflowVersionOptions: Workflow[];
          };
      }
    | {
          value: LaunchState.SELECT_LAUNCH_PLAN;
          context: WorkflowLaunchContext & {
              launchPlanOptions: LaunchPlan[];
              sourceId: NamedEntityIdentifier;
              workflowVersionOptions: Workflow[];
          };
      };

export type TaskLaunchTypestate =
    | BaseLaunchTypestate
    | {
          value: LaunchState.SELECT_TASK_VERSION;
          context: TaskLaunchContext & {
              sourceId: NamedEntityIdentifier;
              taskVersionOptions: Task[];
          };
      };

const defaultBaseContext: BaseLaunchContext = {
    parsedInputs: [],
    showErrors: false,
    unsupportedRequiredInputs: []
};

const defaultHandlers = {
    CANCEL: LaunchState.CANCELLED
};

const baseStateConfig: StatesConfig<
    BaseLaunchContext,
    BaseLaunchSchema,
    BaseLaunchEvent
> = {
    [LaunchState.CANCELLED]: {
        type: 'final'
    },
    [LaunchState.LOADING_INPUTS]: {
        entry: ['hideErrors'],
        invoke: {
            src: 'loadInputs',
            onDone: {
                target: LaunchState.ENTER_INPUTS,
                actions: ['setInputs']
            },
            onError: {
                target: LaunchState.FAILED_LOADING_INPUTS,
                actions: ['setError']
            }
        }
    },
    [LaunchState.FAILED_LOADING_INPUTS]: {
        on: {
            RETRY: LaunchState.LOADING_INPUTS
        }
    },
    [LaunchState.UNSUPPORTED_INPUTS]: {
        // events handled at top level
    },
    [LaunchState.ENTER_INPUTS]: {
        always: {
            target: LaunchState.UNSUPPORTED_INPUTS,
            cond: ({ unsupportedRequiredInputs }) =>
                unsupportedRequiredInputs.length > 0
        },
        on: {
            SUBMIT: LaunchState.SUBMIT_VALIDATING,
            VALIDATE: LaunchState.VALIDATING_INPUTS
        }
    },
    [LaunchState.VALIDATING_INPUTS]: {
        invoke: {
            src: 'validate',
            onDone: LaunchState.ENTER_INPUTS,
            onError: LaunchState.INVALID_INPUTS
        }
    },
    [LaunchState.INVALID_INPUTS]: {
        on: {
            VALIDATE: LaunchState.VALIDATING_INPUTS,
            SUBMIT: LaunchState.SUBMIT_VALIDATING
        }
    },
    [LaunchState.SUBMIT_VALIDATING]: {
        entry: ['showErrors'],
        invoke: {
            src: 'validate',
            onDone: {
                target: LaunchState.SUBMITTING
            },
            onError: {
                target: LaunchState.INVALID_INPUTS,
                actions: ['setError']
            }
        }
    },
    [LaunchState.SUBMITTING]: {
        invoke: {
            src: 'submit',
            onDone: {
                target: LaunchState.SUBMIT_SUCCEEDED,
                actions: ['setExecutionId']
            },
            onError: {
                target: LaunchState.SUBMIT_FAILED,
                actions: ['setError']
            }
        }
    },
    [LaunchState.SUBMIT_FAILED]: {
        on: {
            SUBMIT: LaunchState.SUBMITTING
        }
    },
    [LaunchState.SUBMIT_SUCCEEDED]: {
        type: 'final'
    }
};

export const taskLaunchMachineConfig: MachineConfig<
    TaskLaunchContext,
    TaskLaunchSchema,
    TaskLaunchEvent
> = {
    id: 'launchTask',
    context: { ...defaultBaseContext },
    initial: LaunchState.LOADING_TASK_VERSIONS,
    on: {
        ...defaultHandlers,
        SELECT_TASK_VERSION: {
            target: LaunchState.LOADING_INPUTS,
            actions: ['setTaskVersion']
        }
    },
    states: {
        ...(baseStateConfig as StatesConfig<
            TaskLaunchContext,
            TaskLaunchSchema,
            TaskLaunchEvent
        >),
        [LaunchState.LOADING_TASK_VERSIONS]: {
            invoke: {
                src: 'loadTaskVersions',
                onDone: {
                    target: LaunchState.SELECT_TASK_VERSION,
                    actions: ['setTaskVersionOptions']
                },
                onError: {
                    target: LaunchState.FAILED_LOADING_TASK_VERSIONS,
                    actions: ['setError']
                }
            }
        },
        [LaunchState.FAILED_LOADING_TASK_VERSIONS]: {
            on: {
                RETRY: LaunchState.LOADING_TASK_VERSIONS
            }
        },
        [LaunchState.SELECT_TASK_VERSION]: {
            // events handled at top level
        }
    }
};

export const workflowLaunchMachineConfig: MachineConfig<
    WorkflowLaunchContext,
    WorkflowLaunchSchema,
    WorkflowLaunchEvent
> = {
    id: 'launchWorkflow',
    context: { ...defaultBaseContext },
    initial: LaunchState.LOADING_WORKFLOW_VERSIONS,
    on: {
        ...defaultHandlers,
        SELECT_WORKFLOW_VERSION: {
            target: LaunchState.LOADING_LAUNCH_PLANS,
            actions: ['setWorkflowVersion']
        },
        SELECT_LAUNCH_PLAN: {
            target: LaunchState.LOADING_INPUTS,
            actions: ['setLaunchPlan']
        }
    },
    states: {
        ...(baseStateConfig as StatesConfig<
            WorkflowLaunchContext,
            WorkflowLaunchSchema,
            WorkflowLaunchEvent
        >),
        [LaunchState.LOADING_WORKFLOW_VERSIONS]: {
            invoke: {
                src: 'loadWorkflowVersions',
                onDone: {
                    target: LaunchState.SELECT_WORKFLOW_VERSION,
                    actions: ['setWorkflowVersionOptions']
                },
                onError: {
                    target: LaunchState.FAILED_LOADING_WORKFLOW_VERSIONS,
                    actions: ['setError']
                }
            }
        },
        [LaunchState.FAILED_LOADING_WORKFLOW_VERSIONS]: {
            on: {
                RETRY: LaunchState.LOADING_WORKFLOW_VERSIONS
            }
        },
        [LaunchState.SELECT_WORKFLOW_VERSION]: {
            // Events handled at top level
        },
        [LaunchState.LOADING_LAUNCH_PLANS]: {
            invoke: {
                src: 'loadLaunchPlans',
                onDone: {
                    target: LaunchState.SELECT_LAUNCH_PLAN,
                    actions: ['setLaunchPlanOptions']
                },
                onError: {
                    target: LaunchState.FAILED_LOADING_LAUNCH_PLANS,
                    actions: ['setError']
                }
            }
        },
        [LaunchState.FAILED_LOADING_LAUNCH_PLANS]: {
            on: {
                RETRY: LaunchState.LOADING_LAUNCH_PLANS
            }
        },
        [LaunchState.SELECT_LAUNCH_PLAN]: {
            // events handled at top level
        }
    }
};

type BaseMachineOptions = MachineOptions<BaseLaunchContext, BaseLaunchEvent>;

const baseActions: BaseMachineOptions['actions'] = {
    hideErrors: assign(_ => ({ showErrors: false })),
    setExecutionId: assign((_, event) => ({
        resultExecutionId: (event as ExecutionCreatedEvent).data
    })),
    setInputs: assign((_, event) => {
        const {
            parsedInputs,
            unsupportedRequiredInputs
        } = (event as InputsParsedEvent).data;
        return {
            parsedInputs,
            unsupportedRequiredInputs
        };
    }),
    setError: assign((_, event) => ({
        error: (event as ErrorEvent).data
    })),
    showErrors: assign(_ => ({ showErrors: true }))
};

const baseServices: BaseMachineOptions['services'] = {
    loadInputs: () =>
        Promise.reject('No `loadInputs` service has been provided'),
    submit: () => Promise.reject('No `submit` service has been provided'),
    validate: () => Promise.reject('No `validate` service has been provided')
};

export const taskLaunchMachine = Machine(taskLaunchMachineConfig, {
    actions: {
        ...baseActions,
        setTaskVersion: assign((_, event) => ({
            taskVersion: (event as SelectTaskVersionEvent).taskId
        })),
        setTaskVersionOptions: assign((_, event) => ({
            taskVersionOptions: (event as TaskVersionOptionsLoadedEvent).data
        }))
    },
    services: {
        ...baseServices,
        loadTaskVersions: () =>
            Promise.reject('No `loadTaskVersions` service has been provided')
    }
});

/** A full machine for representing the Launch flow, combining the state definitions
 * with actions/guards/services needed to support them.
 */
export const workflowLaunchMachine = Machine(workflowLaunchMachineConfig, {
    actions: {
        ...baseActions,
        setWorkflowVersion: assign((_, event) => ({
            workflowVersion: (event as SelectWorkflowVersionEvent).workflowId
        })),
        setWorkflowVersionOptions: assign((_, event) => ({
            workflowVersionOptions: (event as DoneInvokeEvent<Workflow[]>).data
        })),
        setLaunchPlanOptions: assign((_, event) => ({
            launchPlanOptions: (event as LaunchPlanOptionsLoadedEvent).data
        })),
        setLaunchPlan: assign((_, event) => ({
            launchPlan: (event as SelectLaunchPlanEvent).launchPlan
        }))
    },
    services: {
        ...baseServices,
        loadWorkflowVersions: () =>
            Promise.reject(
                'No `loadWorkflowVersions` service has been provided'
            ),
        loadLaunchPlans: () =>
            Promise.reject('No `loadLaunchPlans` service has been provided')
    }
});
