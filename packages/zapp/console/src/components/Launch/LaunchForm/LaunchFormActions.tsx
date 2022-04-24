import { Button, DialogActions, FormHelperText } from '@material-ui/core';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import * as React from 'react';
import { history } from 'routes/history';
import { Routes } from 'routes/routes';
import { formStrings } from './constants';
import { LaunchState } from './launchMachine';
import { useStyles } from './styles';
import { BaseInterpretedLaunchState, BaseLaunchService } from './types';

export interface LaunchFormActionsProps {
  state: BaseInterpretedLaunchState;
  service: BaseLaunchService;
  onClose(): void;
}
/** Renders the Submit/Cancel buttons for a LaunchForm */
export const LaunchFormActions: React.FC<LaunchFormActionsProps> = ({
  state,
  service,
  onClose,
}) => {
  const styles = useStyles();
  const submissionInFlight = state.matches(LaunchState.SUBMITTING);
  const canSubmit = [
    LaunchState.ENTER_INPUTS,
    LaunchState.VALIDATING_INPUTS,
    LaunchState.INVALID_INPUTS,
    LaunchState.SUBMIT_FAILED,
  ].some(state.matches);

  const submit: React.FormEventHandler = (event) => {
    event.preventDefault();
    service.send({ type: 'SUBMIT' });
  };

  const onCancel = () => {
    service.send({ type: 'CANCEL' });
    onClose();
  };

  React.useEffect(() => {
    const subscription = service.subscribe((newState) => {
      // On transition to final success state, read the resulting execution
      // id and navigate to the Execution Details page.
      // if (state.matches({ submit: 'succeeded' })) {
      if (newState.matches(LaunchState.SUBMIT_SUCCEEDED)) {
        history.push(Routes.ExecutionDetails.makeUrl(newState.context.resultExecutionId));
      }
    });

    return subscription.unsubscribe;
  }, [service]);

  return (
    <div className={styles.footer}>
      {state.matches(LaunchState.SUBMIT_FAILED) ? (
        <FormHelperText error={true}>{state.context.error.message}</FormHelperText>
      ) : null}
      <DialogActions>
        <Button
          color="primary"
          disabled={submissionInFlight}
          id="launch-workflow-cancel"
          onClick={onCancel}
          variant="outlined"
        >
          {formStrings.cancel}
        </Button>
        <Button
          color="primary"
          disabled={!canSubmit}
          id="launch-workflow-submit"
          onClick={submit}
          type="submit"
          variant="contained"
        >
          {formStrings.submit}
          {submissionInFlight && <ButtonCircularProgress />}
        </Button>
      </DialogActions>
    </div>
  );
};
