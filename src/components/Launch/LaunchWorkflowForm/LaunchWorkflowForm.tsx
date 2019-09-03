import {
    Button,
    DialogActions,
    FormHelperText,
    MenuItem,
    TextField,
    Typography
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { smallFontSize } from 'components/Theme';
import { FilterOperationName, WorkflowId } from 'models';
import * as React from 'react';
import { SearchableSelector } from './SearchableSelector';
import { SimpleInput } from './SimpleInput';
import { InputProps, InputType, LaunchWorkflowFormProps } from './types';
import { UnsupportedInput } from './UnsupportedInput';
import { useLaunchWorkflowFormState } from './useLaunchWorkflowFormState';
import { workflowsToSearchableSelectorOptions } from './utils';

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

function generateFetchSearchResults(
    { listWorkflows }: APIContextValue,
    workflowId: WorkflowId
) {
    return async (query: string) => {
        const { entities: workflows } = await listWorkflows(workflowId, {
            filter: [
                {
                    key: 'version',
                    operation: FilterOperationName.CONTAINS,
                    value: query
                }
            ]
        });
        const options = workflowsToSearchableSelectorOptions(workflows);
        if (options.length > 0) {
            options[0].description = 'latest';
        }
        return options;
    };
}

/** Renders the form for initiating a Launch request based on a Workflow */
export const LaunchWorkflowForm: React.FC<LaunchWorkflowFormProps> = props => {
    const state = useLaunchWorkflowFormState(props);
    const { submissionState } = state;
    const styles = useStyles();
    const fetchSearchResults = generateFetchSearchResults(
        useAPIContext(),
        props.workflowId
    );

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
                <WaitForData
                    spinnerVariant="medium"
                    {...state.workflowOptionsLoadingState}
                >
                    <div className={styles.formControl}>
                        <SearchableSelector
                            label="Workflow Version"
                            onSelectionChanged={state.onSelectWorkflow}
                            options={state.workflowSelectorOptions}
                            fetchSearchResults={fetchSearchResults}
                            selectedItem={state.selectedWorkflow}
                        />
                    </div>

                    <WaitForData
                        spinnerVariant="medium"
                        {...state.inputLoadingState}
                    >
                        {state.inputs.map(input => (
                            <div
                                key={input.label}
                                className={styles.formControl}
                            >
                                {getComponentForInput(input)}
                            </div>
                        ))}
                    </WaitForData>
                </WaitForData>
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
                        disabled={
                            submissionState.loading ||
                            !state.inputLoadingState.hasLoaded
                        }
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
