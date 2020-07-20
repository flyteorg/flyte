import {
    Button,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormHelperText,
    Typography
} from '@material-ui/core';
import { WaitForData } from 'components/common';
import { ButtonCircularProgress } from 'components/common/ButtonCircularProgress';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import {
    FilterOperationName,
    NamedEntityIdentifier,
    SortDirection,
    workflowSortFields
} from 'models';
import * as React from 'react';
import { formStrings, unsupportedRequiredInputsString } from './constants';
import { InputValueCacheContext } from './inputValueCache';
import { LaunchWorkflowFormInputs } from './LaunchWorkflowFormInputs';
import { SearchableSelector } from './SearchableSelector';
import { useStyles } from './styles';
import { LaunchWorkflowFormProps } from './types';
import { UnsupportedRequiredInputsError } from './UnsupportedRequiredInputsError';
import { useLaunchWorkflowFormState } from './useLaunchWorkflowFormState';
import { workflowsToSearchableSelectorOptions } from './utils';

function generateFetchSearchResults(
    { listWorkflows }: APIContextValue,
    workflowId: NamedEntityIdentifier
) {
    return async (query: string) => {
        const { project, domain, name } = workflowId;
        const { entities: workflows } = await listWorkflows(
            { project, domain, name },
            {
                filter: [
                    {
                        key: 'version',
                        operation: FilterOperationName.CONTAINS,
                        value: query
                    }
                ],
                sort: {
                    key: workflowSortFields.createdAt,
                    direction: SortDirection.DESCENDING
                }
            }
        );
        return workflowsToSearchableSelectorOptions(workflows);
    };
}

/** Renders the form for initiating a Launch request based on a Workflow */
export const LaunchWorkflowForm: React.FC<LaunchWorkflowFormProps> = props => {
    const state = useLaunchWorkflowFormState(props);
    const { submissionState } = state;
    const launchPlanSelected = !!state.selectedLaunchPlan;
    const styles = useStyles();
    const fetchSearchResults = generateFetchSearchResults(
        useAPIContext(),
        props.workflowId
    );

    const submit: React.FormEventHandler = event => {
        event.preventDefault();
        state.onSubmit();
    };

    const preventSubmit =
        submissionState.loading ||
        !state.inputLoadingState.hasLoaded ||
        state.unsupportedRequiredInputs.length > 0;

    return (
        <InputValueCacheContext.Provider value={state.inputValueCache}>
            <DialogTitle disableTypography={true} className={styles.header}>
                <div className={styles.inputLabel}>{formStrings.title}</div>
                <Typography variant="h6">{state.workflowName}</Typography>
            </DialogTitle>
            <DialogContent dividers={true} className={styles.inputsSection}>
                <WaitForData
                    spinnerVariant="medium"
                    {...state.workflowOptionsLoadingState}
                >
                    <section
                        title={formStrings.workflowVersion}
                        className={styles.formControl}
                    >
                        <SearchableSelector
                            id="launch-workflow-selector"
                            label={formStrings.workflowVersion}
                            onSelectionChanged={state.onSelectWorkflow}
                            options={state.workflowSelectorOptions}
                            fetchSearchResults={fetchSearchResults}
                            selectedItem={state.selectedWorkflow}
                        />
                    </section>
                    <WaitForData
                        {...state.launchPlanOptionsLoadingState}
                        spinnerVariant="medium"
                    >
                        <section
                            title={formStrings.launchPlan}
                            className={styles.formControl}
                        >
                            <SearchableSelector
                                id="launch-lp-selector"
                                label={formStrings.launchPlan}
                                onSelectionChanged={state.onSelectLaunchPlan}
                                options={state.launchPlanSelectorOptions}
                                selectedItem={state.selectedLaunchPlan}
                            />
                        </section>
                    </WaitForData>
                    {launchPlanSelected ? (
                        <WaitForData
                            spinnerVariant="medium"
                            {...state.inputLoadingState}
                        >
                            <section title={formStrings.inputs}>
                                {state.unsupportedRequiredInputs.length > 0 ? (
                                    <UnsupportedRequiredInputsError
                                        inputs={state.unsupportedRequiredInputs}
                                    />
                                ) : (
                                    <LaunchWorkflowFormInputs
                                        key={state.formKey}
                                        inputs={state.inputs}
                                        ref={state.formInputsRef}
                                        showErrors={state.showErrors}
                                    />
                                )}
                            </section>
                        </WaitForData>
                    ) : null}
                </WaitForData>
            </DialogContent>
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
                        {formStrings.cancel}
                    </Button>
                    <Button
                        color="primary"
                        disabled={preventSubmit}
                        id="launch-workflow-submit"
                        onClick={submit}
                        type="submit"
                        variant="contained"
                    >
                        {formStrings.submit}
                        {submissionState.loading && <ButtonCircularProgress />}
                    </Button>
                </DialogActions>
            </div>
        </InputValueCacheContext.Provider>
    );
};
