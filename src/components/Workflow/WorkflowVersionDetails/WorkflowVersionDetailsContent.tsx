import { SectionHeader, ViewHeader } from 'components/common';
import { Workflow } from 'models';
import * as React from 'react';

export interface WorkflowVersionDetailsContentProps {
    workflow: Workflow;
}

export const WorkflowVersionDetailsContent: React.FC<WorkflowVersionDetailsContentProps> = ({
    workflow
}) => {
    const { id } = workflow;
    return (
        <>
            <ViewHeader title={id.name} subtitle={id.version} />
            <SectionHeader title="Workflow Graph" />
        </>
    );
};
