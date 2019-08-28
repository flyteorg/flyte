import {
    Button,
    DialogActions,
    FormHelperText,
    Typography
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { rejectAfter } from 'common/promiseUtils';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import { FetchFn } from 'components/hooks';
import { smallFontSize } from 'components/Theme';
import * as React from 'react';
import { SimpleInput } from './SimpleInput';
import { InputProps, InputType, LaunchWorkflowFormProps } from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { useLaunchWorkflowFormState } from './useLaunchWorkflowFormState';
import { WorkflowSelector, WorkflowSelectorOption } from './WorkflowSelector';

const useStyles = makeStyles((theme: Theme) => ({
    footer: {
        borderTop: `1px solid ${theme.palette.divider}`,
        padding: theme.spacing(2)
    },
    formControl: {
        padding: `${theme.spacing(1.5)}px 0`
    },
    header: {
        borderBottom: `1px solid ${theme.palette.divider}`,
        padding: theme.spacing(2),
        width: '100%'
    },
    inputsSection: {
        padding: theme.spacing(2)
    },
    inputLabel: {
        color: theme.palette.text.hint,
        fontSize: smallFontSize
    },
    root: {
        display: 'flex',
        flexDirection: 'column',
        width: '100%'
    }
}));

function getComponentForInput(input: InputProps) {
    switch (input.typeDefinition.type) {
        case InputType.Collection:
        case InputType.Map:
        case InputType.Schema:
        case InputType.Unknown:
        case InputType.None:
            return <UnsupportedInput {...input} />;
        default:
            return <SimpleInput {...input} />;
    }
}

const fetchSearchResults: FetchFn<WorkflowSelectorOption[], string> = () =>
    rejectAfter(0, 'Not Implemented');

/** Renders the form for initiating a Launch request based on a Workflow */
export const LaunchWorkflowForm: React.FC<LaunchWorkflowFormProps> = props => {
    const state = useLaunchWorkflowFormState(props);
    const { submissionState } = state;
    const styles = useStyles();
    const submit: React.FormEventHandler = event => {
        event.preventDefault();
        state.onSubmit();
    };

    return (
        <form className={styles.root}>
            <header className={styles.header}>
                <div className={styles.inputLabel}>Launch Workflow</div>
                <Typography variant="h6">{state.workflowName}</Typography>
            </header>
            <section className={styles.inputsSection}>
                <WorkflowSelector
                    onSelectionChanged={state.onSelectWorkflow}
                    options={state.workflowSelectorOptions}
                    fetchSearchResults={fetchSearchResults}
                    selectedItem={state.selectedWorkflow}
                />
                {state.inputs.map(input => (
                    <div key={input.label} className={styles.formControl}>
                        {getComponentForInput(input)}
                    </div>
                ))}
            </section>
            <div className={styles.footer}>
                {!!submissionState.lastError && (
                    <FormHelperText error={true}>
                        {submissionState.lastError.message}
                    </FormHelperText>
                )}
                <DialogActions>
                    <Button
                        color="primary"
                        disabled={submissionState.loading}
                        id="launch-workflow-cancel"
                        onClick={state.onCancel}
                        variant="outlined"
                    >
                        Cancel
                    </Button>
                    <Button
                        color="primary"
                        disabled={submissionState.loading}
                        id="launch-workflow-submit"
                        onClick={submit}
                        type="submit"
                        variant="contained"
                    >
                        Launch
                        {submissionState.loading && <ButtonCircularProgress />}
                    </Button>
                </DialogActions>
            </div>
        </form>
    );
};
