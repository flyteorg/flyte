import { assign, Machine, State, StateNodeConfig } from 'xstate';
import {
  FetchEventObject,
  fetchEvents,
  FetchMachine,
  FetchStateContext,
  fetchStates,
  FetchStateSchema,
} from './types';

/** Helper function to determine if the State object from a fetchMachine
 * is one in which data is being fetched.
 */
export function isLoadingState(state: State<any, any>) {
  return (
    state.matches(fetchStates.LOADING) ||
    state.matches(fetchStates.FAILED_RETRYING) ||
    state.matches(fetchStates.REFRESHING) ||
    state.matches(fetchStates.REFRESH_FAILED_RETRYING)
  );
}

function makeLoadConfig<T>(
  successTarget: string,
  failureTarget: string,
): Partial<StateNodeConfig<FetchStateContext<T>, FetchStateSchema, FetchEventObject>> {
  return {
    entry: assign({ lastError: (_) => null }),
    invoke: {
      src: 'doFetch',
      onDone: {
        target: successTarget,
        actions: assign({
          value: (_, event) => event.data,
        }),
      },
      onError: {
        target: failureTarget,
        actions: assign({ lastError: (_, event) => event.data }),
      },
    },
  };
}

const defaultContext: Partial<FetchStateContext<unknown>> = {
  lastError: null,
};

/** A generic state machine for representing fetchable data. */
export const fetchMachine: FetchMachine<unknown> = Machine<
  FetchStateContext<unknown>,
  FetchEventObject
>({
  id: 'fetch',
  initial: fetchStates.IDLE,
  states: {
    [fetchStates.IDLE]: {
      entry: [
        // When resetting, clear the context back to an initial state
        assign((context) => ({
          ...defaultContext,
          value: context.defaultValue,
        })),
      ],
      on: {
        [fetchEvents.RESET]: `#fetch.${fetchStates.IDLE}`,
        [fetchEvents.LOAD]: `#fetch.${fetchStates.LOADING}`,
      },
    },
    [fetchStates.LOADING]: makeLoadConfig(
      `#fetch.${fetchStates.LOADED}`,
      `#fetch.${fetchStates.FAILED}`,
    ),
    [fetchStates.LOADED]: {
      on: {
        [fetchEvents.RESET]: `#fetch.${fetchStates.IDLE}`,
        [fetchEvents.LOAD]: `#fetch.${fetchStates.REFRESHING}`,
      },
    },
    [fetchStates.FAILED]: {
      on: {
        [fetchEvents.RESET]: `#fetch.${fetchStates.IDLE}`,
        [fetchEvents.LOAD]: `#fetch.${fetchStates.FAILED_RETRYING}`,
      },
    },
    [fetchStates.FAILED_RETRYING]: makeLoadConfig(
      `#fetch.${fetchStates.LOADED}`,
      `#fetch.${fetchStates.FAILED}`,
    ),
    [fetchStates.REFRESHING]: makeLoadConfig(
      `#fetch.${fetchStates.LOADED}`,
      `#fetch.${fetchStates.REFRESH_FAILED}`,
    ),
    [fetchStates.REFRESH_FAILED]: {
      on: {
        [fetchEvents.RESET]: `#fetch.${fetchStates.IDLE}`,
        [fetchEvents.LOAD]: `#fetch.${fetchStates.REFRESH_FAILED_RETRYING}`,
      },
    },
    [fetchStates.REFRESH_FAILED_RETRYING]: makeLoadConfig(
      `#fetch.${fetchStates.LOADED}`,
      `#fetch.${fetchStates.REFRESH_FAILED}`,
    ),
  },
});
