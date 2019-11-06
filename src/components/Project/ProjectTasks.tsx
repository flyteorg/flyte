import { Typography } from '@material-ui/core';
import { SearchResult, WaitForData } from 'components/common';
import {
    SearchableNamedEntity,
    SearchableNamedEntityList,
    useNamedEntityListStyles
} from 'components/common/SearchableNamedEntityList';
import { useCommonStyles } from 'components/common/styles';
import { useTaskNameList } from 'components/hooks/useNamedEntity';
import { limits, SortDirection, workflowSortFields } from 'models';
import * as React from 'react';

export interface ProjectTasksProps {
    projectId: string;
    domainId: string;
}

/** A listing of the Tasks registered for a project */
export const ProjectTasks: React.FC<ProjectTasksProps> = ({
    domainId: domain,
    projectId: project
}) => {
    const listStyles = useNamedEntityListStyles();
    const commonStyles = useCommonStyles();
    const renderItem = ({
        key,
        value,
        content
    }: SearchResult<SearchableNamedEntity>) => (
        <li key={key}>
            <div className={listStyles.searchResult}>
                <div className={listStyles.itemName}>
                    <div>{content}</div>
                    {!!value.metadata.description && (
                        <Typography
                            variant="body2"
                            className={commonStyles.hintText}
                        >
                            {value.metadata.description}
                        </Typography>
                    )}
                </div>
            </div>
        </li>
    );
    const taskNames = useTaskNameList(
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
        <WaitForData {...taskNames}>
            <SearchableNamedEntityList
                names={taskNames.value}
                renderItem={renderItem}
            />
        </WaitForData>
    );
};
