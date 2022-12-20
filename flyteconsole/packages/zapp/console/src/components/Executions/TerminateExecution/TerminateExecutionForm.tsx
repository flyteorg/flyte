import { FormControl, FormLabel, OutlinedInput } from '@material-ui/core';
import Button from '@material-ui/core/Button';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import { useCommonStyles } from 'components/common/styles';
import { smallFontSize } from 'components/Theme/constants';
import * as React from 'react';
import { useTerminateExecutionState } from './useTerminateExecutionState';

const useStyles = makeStyles((theme: Theme) => ({
  buttonGroup: {
    justifyContent: 'center',
  },
  input: {
    fontSize: smallFontSize,
  },
  root: {
    width: theme.spacing(30),
    padding: theme.spacing(2),
  },
  title: {
    marginBottom: theme.spacing(1),
    textTransform: 'uppercase',
    color: theme.palette.text.secondary,
  },
}));

// This corresponds to the maximum length allowed by the API.
const defaultCauseString = 'Terminated from UI';
const placeholderString = 'Reason for termination (optional)';

/** A small form for creating and submitting a request to terminate a workflow
 * execution. This includes error/retry logic in the case of an API failure.
 */
export const TerminateExecutionForm: React.FC<{
  onClose: (...args: any) => void;
}> = ({ onClose }) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const {
    cause,
    setCause,
    terminationState: { error, isLoading: terminating },
    terminateExecution,
  } = useTerminateExecutionState(onClose);

  const onChange: React.ChangeEventHandler<HTMLInputElement> = ({ target: { value } }) =>
    setCause(value);

  const submit: React.FormEventHandler = (event) => {
    event.preventDefault();
    terminateExecution(cause || defaultCauseString);
  };

  return (
    <form className={styles.root}>
      <FormLabel component="legend" className={styles.title}>
        Terminate Workflow
      </FormLabel>
      <FormControl margin="dense" variant="outlined" fullWidth={true}>
        <OutlinedInput
          autoFocus={true}
          className={styles.input}
          fullWidth={true}
          labelWidth={0}
          multiline={true}
          onChange={onChange}
          placeholder={placeholderString}
          rowsMax={4}
          rows={4}
          type="text"
          value={cause}
        />
      </FormControl>
      {error && <p className={commonStyles.errorText}>{`${error}`}</p>}
      <div className={commonStyles.formButtonGroup}>
        <Button
          color="primary"
          disabled={terminating}
          onClick={submit}
          type="submit"
          variant="contained"
        >
          Terminate
          {terminating && <ButtonCircularProgress />}
        </Button>
        <Button
          color="primary"
          disabled={terminating}
          id="terminate-execution-cancel"
          onClick={onClose}
          variant="outlined"
        >
          Cancel
        </Button>
      </div>
    </form>
  );
};
