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
import { ExecutionNodeViews } from './ExecutionNodeViews';

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
