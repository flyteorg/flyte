import { Collapse, IconButton } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import ExpandMore from '@material-ui/icons/ExpandMore';
import * as classnames from 'classnames';
import { WaitForData, withRouteParams } from 'components/common';
import { RefreshConfig, useDataRefresher } from 'components/hooks';
import { Execution } from 'models';
import * as React from 'react';
import { executionRefreshIntervalMs } from '../constants';
import { ExecutionContext, ExecutionDataCacheContext } from '../contexts';
import { useExecutionDataCache } from '../useExecutionDataCache';
import { useWorkflowExecution } from '../useWorkflowExecution';
import { executionIsTerminal } from '../utils';
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

const executionRefreshConfig: RefreshConfig<Execution> = {
    interval: executionRefreshIntervalMs,
    valueIsFinal: executionIsTerminal
};

/** The view component for the Execution Details page */
export const ExecutionDetailsContainer: React.FC<ExecutionDetailsRouteParams> = ({
    executionId,
    domainId,
    projectId
}) => {
    const id = {
        project: projectId,
        domain: domainId,
        name: executionId
    };
    const styles = useStyles();
    const [metadataExpanded, setMetadataExpanded] = React.useState(true);
    const toggleMetadata = () => setMetadataExpanded(!metadataExpanded);
    const dataCache = useExecutionDataCache();
    const { fetchable, terminateExecution } = useWorkflowExecution(
        id,
        dataCache
    );
    useDataRefresher(id, fetchable, executionRefreshConfig);
    const contextValue = {
        terminateExecution,
        execution: fetchable.value
    };

    return (
        <WaitForData {...fetchable}>
            <ExecutionContext.Provider value={contextValue}>
                <ExecutionDetailsAppBarContent execution={fetchable.value} />
                <div className={styles.metadataContainer}>
                    <Collapse in={metadataExpanded}>
                        <ExecutionMetadata execution={fetchable.value} />
                    </Collapse>
                    <div className={styles.expandCollapseContainer}>
                        <IconButton size="small" onClick={toggleMetadata}>
                            <ExpandMore
                                className={classnames(
                                    styles.expandCollapseButton,
                                    {
                                        expanded: metadataExpanded
                                    }
                                )}
                            />
                        </IconButton>
                    </div>
                </div>
                <ExecutionDataCacheContext.Provider value={dataCache}>
                    <ExecutionNodeViews execution={fetchable.value} />
                </ExecutionDataCacheContext.Provider>
            </ExecutionContext.Provider>
        </WaitForData>
    );
};

export const ExecutionDetails = withRouteParams<ExecutionDetailsRouteParams>(
    ExecutionDetailsContainer
);
