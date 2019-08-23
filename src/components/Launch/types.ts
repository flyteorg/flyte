import { FetchableData } from 'components/hooks';
import {
    LaunchPlan,
    Workflow,
    WorkflowExecutionIdentifier,
    WorkflowId
} from 'models';

export interface LaunchWorkflowFormProps {
    launchPlan?: LaunchPlan;
    workflowId: WorkflowId;
    workflow?: Workflow;
}

export interface LaunchWorkflowFormState {
    inputs: InputProps[];
    launchPlan?: LaunchPlan;
    submissionState: FetchableData<WorkflowExecutionIdentifier>;
    workflow?: Workflow;
    workflowName: string;
    onCancel(): void;
    onSubmit(): void;
    setLaunchPlan(launchPlan: LaunchPlan): void;
    setWorkflow(workflow: Workflow): void;
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
