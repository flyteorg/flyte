import { makeStyles, Theme } from '@material-ui/core/styles';
import { SectionHeader, WaitForData } from 'components/common';
import { WorkflowExecutionsTable } from 'components/Executions/Tables/WorkflowExecutionsTable';
import { useWorkflowExecutions } from 'components/hooks';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { SortDirection } from 'models/AdminEntity';
import { executionSortFields } from 'models/Execution';
import * as React from 'react';
import { recentExecutionsLimit } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        flexDirection: 'column',
        height: theme.spacing(35)
    }
}));

interface ProjectRecentExecutionsProps {
    domain: string;
    project: string;
}

export const ProjectRecentExecutions: React.FC<ProjectRecentExecutionsProps> = ({
    domain,
    project
}) => {
    const styles = useStyles();

    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.DESCENDING
    };
    const limit = recentExecutionsLimit;

    const executions = useWorkflowExecutions(
        { domain, project },
        { limit, sort }
    );

    return (
        <section>
            <SectionHeader title="Recent Executions" />
            <div className={styles.container}>
                <WaitForData {...executions}>
                    <WorkflowExecutionsTable
                        {...executions}
                        isFetching={isLoadingState(executions.state)}
                        // This is a simple quick-glance list, we don't want to
                        // do pagination here
                        moreItemsAvailable={false}
                    />
                </WaitForData>
            </div>
        </section>
    );
};
