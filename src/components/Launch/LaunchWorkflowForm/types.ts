import { FetchableData, MultiFetchableState } from 'components/hooks';
import { LaunchPlan, WorkflowExecutionIdentifier, WorkflowId } from 'models';
import { SearchableSelectorOption } from './SearchableSelector';

export interface LaunchWorkflowFormProps {
    workflowId: WorkflowId;
}

export interface LaunchWorkflowFormState {
    defaultLaunchPlan: FetchableData<LaunchPlan | undefined>;
    inputLoadingState: MultiFetchableState;
    inputs: InputProps[];
    launchPlans: FetchableData<LaunchPlan[]>;
    selectedLaunchPlan?: LaunchPlan;
    submissionState: FetchableData<WorkflowExecutionIdentifier>;
    selectedWorkflow?: SearchableSelectorOption<WorkflowId>;
    workflowName: string;
    workflowOptionsLoadingState: MultiFetchableState;
    workflowSelectorOptions: SearchableSelectorOption<WorkflowId>[];
    onCancel(): void;
    onSelectWorkflow(selected: SearchableSelectorOption<WorkflowId>): void;
    onSubmit(): void;
    setLaunchPlan(launchPlan: LaunchPlan): void;
}

export enum InputType {
    Binary = 'BINARY',
    Blob = 'BLOB',
    Boolean = 'BOOLEAN',
    Collection = 'COLLECTION',
    Datetime = 'DATETIME',
    Duration = 'DURATION',
    Error = 'ERROR',
    Float = 'FLOAT',
    Integer = 'INTEGER',
    Map = 'MAP',
    None = 'NONE',
    Schema = 'SCHEMA',
    String = 'STRING',
    Struct = 'STRUCT',
    Unknown = 'UNKNOWN'
}

export interface InputTypeDefinition {
    type: InputType;
    subtype?: InputTypeDefinition;
}

export type InputValue = string | number | boolean | Date;
export type InputChangeHandler = (newValue: InputValue) => void;

export interface InputProps {
    description: string;
    error?: string;
    helperText?: string;
    name: string;
    label: string;
    required: boolean;
    typeDefinition: InputTypeDefinition;
    value?: InputValue;
    onChange: InputChangeHandler;
}
