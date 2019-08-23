import { Button, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import { useCommonStyles } from 'components/common/styles';
import { Execution } from 'models';
import * as React from 'react';
import { useRelaunchExecutionState } from './useRelaunchExecutionState';

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        width: theme.spacing(30),
        padding: theme.spacing(2)
    },
    text: {
        marginBottom: theme.spacing(2)
    }
}));

/** A small form for confirming a request to relaunch an execution. */
export const RelaunchExecutionForm: React.FC<{
    execution: Execution;
    onClose: (...args: any) => void;
}> = ({ execution, onClose }) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const {
        fetchable: { loading, lastError },
        relaunchWorkflowExecution
    } = useRelaunchExecutionState(execution, onClose);

    const submit: React.FormEventHandler = event => {
        event.preventDefault();
        relaunchWorkflowExecution();
    };

    return (
        <form className={styles.root}>
            <Typography className={styles.text} variant="body1">
                This will relaunch the execution using the same inputs.
                {lastError && (
                    <div className={commonStyles.errorText}>{`${
                        lastError.message
                    }`}</div>
                )}
            </Typography>
            <div className={commonStyles.formButtonGroup}>
                <Button
                    color="primary"
                    disabled={loading}
                    onClick={submit}
                    type="submit"
                    variant="contained"
                >
                    Confirm
                    {loading && <ButtonCircularProgress />}
                </Button>
                <Button
                    color="primary"
                    disabled={loading}
                    id="relaunch-execution-cancel"
                    onClick={onClose}
                    variant="outlined"
                >
                    Cancel
                </Button>
            </div>
        </form>
    );
};
