import { DialogContent } from '@material-ui/core';
import { getCacheKey } from 'components/Cache/utils';
import * as React from 'react';
import { formStrings } from './constants';
import { LaunchFormActions } from './LaunchFormActions';
import { LaunchFormHeader } from './LaunchFormHeader';
import { LaunchFormInputs } from './LaunchFormInputs';
import { LaunchState } from './launchMachine';
import { LaunchRoleInput } from './LaunchRoleInput';
import { SearchableSelector } from './SearchableSelector';
import { useStyles } from './styles';
import { BaseInterpretedLaunchState, BaseLaunchService, LaunchTaskFormProps } from './types';
import { useLaunchTaskFormState } from './useLaunchTaskFormState';
import { isEnterInputsState } from './utils';

/** Renders the form for initiating a Launch request based on a Task */
export const LaunchTaskForm: React.FC<LaunchTaskFormProps> = (props) => {
  const { formInputsRef, roleInputRef, state, service, taskSourceSelectorState } =
    useLaunchTaskFormState(props);
  const styles = useStyles();
  const baseState = state as BaseInterpretedLaunchState;
  const baseService = service as BaseLaunchService;

  // Any time the inputs change (even if it's just re-ordering), we must
  // change the form key so that the inputs component will re-mount.
  const formKey = React.useMemo<string>(() => {
    return getCacheKey(state.context.parsedInputs);
  }, [state.context.parsedInputs]);

  const { fetchSearchResults, onSelectTaskVersion, selectedTask, taskSelectorOptions } =
    taskSourceSelectorState;

  const showTaskSelector = ![
    LaunchState.LOADING_TASK_VERSIONS,
    LaunchState.FAILED_LOADING_TASK_VERSIONS,
  ].some(state.matches);

  // TODO: We removed all loading indicators here. Decide if we want skeletons
  // instead.
  // https://github.com/lyft/flyte/issues/666

  return (
    <>
      <LaunchFormHeader title={state.context.sourceId?.name} />
      <DialogContent dividers={true} className={styles.inputsSection}>
        {showTaskSelector ? (
          <section title={formStrings.taskVersion} className={styles.formControl}>
            <SearchableSelector
              id="launch-task-selector"
              label={formStrings.taskVersion}
              onSelectionChanged={onSelectTaskVersion}
              options={taskSelectorOptions}
              fetchSearchResults={fetchSearchResults}
              selectedItem={selectedTask}
            />
          </section>
        ) : null}
        {isEnterInputsState(baseState) ? (
          <LaunchRoleInput
            initialValue={state.context.defaultAuthRole}
            ref={roleInputRef}
            showErrors={state.context.showErrors}
          />
        ) : null}
        <LaunchFormInputs key={formKey} ref={formInputsRef} state={baseState} variant="task" />
      </DialogContent>
      <LaunchFormActions state={baseState} service={baseService} onClose={props.onClose} />
    </>
  );
};
