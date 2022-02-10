import { Collapse, IconButton } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ExpandMore from '@material-ui/icons/ExpandMore';
import classnames from 'classnames';
import { LargeLoadingSpinner } from 'components/common/LoadingSpinner';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { withRouteParams } from 'components/common/withRouteParams';
import { DataError } from 'components/Errors/DataError';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { ExecutionContext } from '../contexts';
import { useWorkflowExecutionQuery } from '../useWorkflowExecution';
import { ExecutionDetailsAppBarContent } from './ExecutionDetailsAppBarContent';
import { ExecutionMetadata } from './ExecutionMetadata';
import { ExecutionNodeViews } from './ExecutionNodeViews';

const useStyles = makeStyles((theme: Theme) => ({
    expandCollapseButton: {
        transition: theme.transitions.create('transform'),
        '&.expanded': {
            transform: 'rotate(180deg)'
        }
    },
    expandCollapseContainer: {
        alignItems: 'center',
        bottom: 0,
        display: 'flex',
        // Matches height of tabs in the NodeViews container
        height: theme.spacing(6),
        position: 'absolute',
        right: theme.spacing(3),
        transform: 'translateY(100%)',
        zIndex: 1
    },
    metadataContainer: {
        position: 'relative'
    }
}));

export interface ExecutionDetailsRouteParams {
    domainId: string;
    executionId: string;
    projectId: string;
}
export type ExecutionDetailsProps = ExecutionDetailsRouteParams;

interface RenderExecutionDetailsProps {
    execution: Execution;
}

const RenderExecutionDetails: React.FC<RenderExecutionDetailsProps> = ({
    execution
}) => {
    const styles = useStyles();
    const [metadataExpanded, setMetadataExpanded] = React.useState(true);
    const toggleMetadata = () => setMetadataExpanded(!metadataExpanded);
    const contextValue = {
        execution
    };

    return (
        <ExecutionContext.Provider value={contextValue}>
            <ExecutionDetailsAppBarContent execution={execution} />
            <div className={styles.metadataContainer}>
                <Collapse in={metadataExpanded}>
                    <ExecutionMetadata execution={execution} />
                </Collapse>
                <div className={styles.expandCollapseContainer}>
                    <IconButton size="small" onClick={toggleMetadata}>
                        <ExpandMore
                            className={classnames(styles.expandCollapseButton, {
                                expanded: metadataExpanded
                            })}
                        />
                    </IconButton>
                </div>
            </div>
            <ExecutionNodeViews execution={execution} />
        </ExecutionContext.Provider>
    );
};

/** The view component for the Execution Details page */
export const ExecutionDetailsContainer: React.FC<ExecutionDetailsProps> = ({
    executionId,
    domainId,
    projectId
}) => {
    const id = {
        project: projectId,
        domain: domainId,
        name: executionId
    };

    const renderExecutionDetails = (execution: Execution) => (
        <RenderExecutionDetails execution={execution} />
    );

    return (
        <WaitForQuery
            errorComponent={DataError}
            loadingComponent={LargeLoadingSpinner}
            query={useWorkflowExecutionQuery(id)}
        >
            {renderExecutionDetails}
        </WaitForQuery>
    );
};

export const ExecutionDetails = withRouteParams<ExecutionDetailsRouteParams>(
    ExecutionDetailsContainer
);
