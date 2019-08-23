import { TextField } from '@material-ui/core';
import { FetchableData } from 'components/hooks';
import { NamedEntityIdentifier, WorkflowId } from 'models';
import * as React from 'react';

export interface WorkflowSelectorOption {
    id: string;
    data: WorkflowId;
    name: string;
    description?: string;
}

export interface WorkflowSelectorProps {
    options: WorkflowSelectorOption[];
    searchResults: FetchableData<WorkflowSelectorOption[]>;
    searchValue?: string;
    selectedItem?: WorkflowSelectorOption;
    workflowId: NamedEntityIdentifier;
    onSearchStringChanged(newValue: string): void;
    onSelectionChanged(newSelection: WorkflowSelectorOption): void;
}

export const WorkflowSelector: React.FC<WorkflowSelectorProps> = ({
    options,
    searchResults,
    searchValue = '',
    selectedItem,
    workflowId,
    onSearchStringChanged,
    onSelectionChanged
}) => {
    const searchActive = searchValue.length > 0;
    let inputValue = '';

    if (searchActive) {
        inputValue = searchValue;
    } else if (selectedItem) {
        inputValue = selectedItem.name;
    }

    const onChange = ({
        target: { value }
    }: React.ChangeEvent<HTMLInputElement>) => onSearchStringChanged(value);

    return (
        <div>
            <TextField
                label="Workflow Version"
                onChange={onChange}
                value={inputValue}
            />
        </div>
    );
};
