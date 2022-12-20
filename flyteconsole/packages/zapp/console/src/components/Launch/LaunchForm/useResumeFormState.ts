import { useMachine } from '@xstate/react';
import { defaultStateMachineConfig } from 'components/common/constants';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { Core } from 'flyteidl';
import { partial } from 'lodash';
import { CompiledNode } from 'models/Node/types';
import { RefObject, useMemo, useRef } from 'react';
import {
  TaskResumeContext,
  TaskResumeTypestate,
  taskResumeMachine,
  BaseLaunchEvent,
} from './launchMachine';
import { validate as baseValidate } from './services';
import {
  BaseLaunchFormProps,
  LaunchFormInputsRef,
  ParsedInput,
  ResumeFormState,
  TaskInitialLaunchParameters,
} from './types';
import { getInputDefintionForLiteralType, getUnsupportedRequiredInputs } from './utils';

interface ResumeFormProps extends BaseLaunchFormProps {
  compiledNode: CompiledNode;
  initialParameters?: TaskInitialLaunchParameters;
  nodeId: string;
}

async function loadInputs({ compiledNode }: TaskResumeContext) {
  if (!compiledNode) {
    throw new Error('Failed to load inputs: missing compiledNode');
  }
  const signalType = compiledNode.gateNode?.signal?.type;
  if (!signalType) {
    throw new Error('Failed to load inputs: missing signal.type');
  }
  const parsedInputs: ParsedInput[] = [
    {
      description: '',
      label: 'Signal Input',
      name: 'signal',
      required: true,
      typeDefinition: getInputDefintionForLiteralType({ simple: signalType.simple ?? undefined }),
    },
  ];

  return {
    parsedInputs,
    unsupportedRequiredInputs: getUnsupportedRequiredInputs(parsedInputs),
  };
}

async function validate(formInputsRef: RefObject<LaunchFormInputsRef>) {
  return baseValidate(formInputsRef);
}

async function submit(
  { resumeSignalNode }: APIContextValue,
  formInputsRef: RefObject<LaunchFormInputsRef>,
  { compiledNode }: TaskResumeContext,
) {
  if (!compiledNode?.gateNode?.signal?.signalId) {
    throw new Error('SignalId is empty');
  }
  if (formInputsRef.current === null) {
    throw new Error('Unexpected empty form inputs ref');
  }

  const literals = formInputsRef.current.getValues();

  const response = await resumeSignalNode({
    id: compiledNode?.gateNode?.signal?.signalId as unknown as Core.SignalIdentifier,
    value: literals['signal'],
  });

  return response;
}

function getServices(apiContext: APIContextValue, formInputsRef: RefObject<LaunchFormInputsRef>) {
  return {
    loadInputs: partial(loadInputs),
    submit: partial(submit, apiContext, formInputsRef),
    validate: partial(validate, formInputsRef),
  };
}

/** Contains all of the form state for a LaunchTaskForm, including input
 * definitions, current input values, and errors.
 */
export function useResumeFormState({ compiledNode }: ResumeFormProps): ResumeFormState {
  const apiContext = useAPIContext();
  const formInputsRef = useRef<LaunchFormInputsRef>(null);

  const services = useMemo(
    () => getServices(apiContext, formInputsRef),
    [apiContext, formInputsRef],
  );

  const [state, , service] = useMachine<TaskResumeContext, BaseLaunchEvent, TaskResumeTypestate>(
    taskResumeMachine,
    {
      ...defaultStateMachineConfig,
      services,
      context: {
        compiledNode,
      },
    },
  );

  return {
    formInputsRef,
    state,
    service,
  };
}
