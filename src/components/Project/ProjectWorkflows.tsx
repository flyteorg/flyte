import { Typography } from '@material-ui/core';
import ChevronRight from '@material-ui/icons/ChevronRight';
import { SearchResult, WaitForData } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { useWorkflowNameList } from 'components/hooks/useNamedEntity';
import {
    SearchableNamedEntity,
    SearchableNamedEntityList,
    useNamedEntityListStyles
} from 'components/Workflow/SearchableNamedEntityList';
import { SearchableWorkflowNameList } from 'components/Workflow/SearchableWorkflowNameList';
import { limits, SortDirection, workflowSortFields } from 'models';
import * as React from 'react';
import { Link } from 'react-router-dom';
import { Routes } from 'routes';

export interface ProjectWorkflowsProps {
    projectId: string;
    domainId: string;
}

/** A listing of the Workflows registered for a project */
export const ProjectWorkflows: React.FC<ProjectWorkflowsProps> = ({
    domainId: domain,
    projectId: project
}) => {
    const workflowNames = useWorkflowNameList(
        { domain, project },
        {
            limit: limits.NONE,
            sort: {
                direction: SortDirection.ASCENDING,
                key: workflowSortFields.name
            }
        }
    );

    return (
        <WaitForData {...workflowNames}>
            <SearchableWorkflowNameList names={workflowNames.value} />
        </WaitForData>
    );
};
