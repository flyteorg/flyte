import classnames from 'classnames';
import { Admin } from 'flyteidl';
import { getCacheKey } from 'components/Cache/utils';
import { DetailsPanel } from 'components/common/DetailsPanel';
import { useCommonStyles } from 'components/common/styles';
import * as scrollbarSize from 'dom-helpers/util/scrollbarSize';
import { NodeExecution, NodeExecutionIdentifier } from 'models/Execution/types';
import * as React from 'react';
import { NodeExecutionDetailsPanelContent } from '../ExecutionDetails/NodeExecutionDetailsPanelContent';
import { NodeExecutionsTableContext } from './contexts';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { generateColumns } from './nodeExecutionColumns';
import { NodeExecutionRow } from './NodeExecutionRow';
import { NoExecutionsContent } from './NoExecutionsContent';
import { useColumnStyles, useExecutionTableStyles } from './styles';

export interface NodeExecutionsTableProps {
    abortMetadata?: Admin.IAbortMetadata;
    nodeExecutions: NodeExecution[];
}

const scrollbarPadding = scrollbarSize();

/** Renders a table of NodeExecution records. Executions with errors will
 * have an expanadable container rendered as part of the table row.
 * NodeExecutions are expandable and will potentially render a list of child
 * TaskExecutions
 */
export const NodeExecutionsTable: React.FC<NodeExecutionsTableProps> = ({
    abortMetadata,
    nodeExecutions
}) => {
    const [
        selectedExecution,
        setSelectedExecution
    ] = React.useState<NodeExecutionIdentifier | null>(null);
    const commonStyles = useCommonStyles();
    const tableStyles = useExecutionTableStyles();

    const executionsWithKeys = React.useMemo(
        () =>
            nodeExecutions.map(nodeExecution => ({
                nodeExecution,
                cacheKey: getCacheKey(nodeExecution.id)
            })),
        [nodeExecutions]
    );

    const columnStyles = useColumnStyles();
    // Memoizing columns so they won't be re-generated unless the styles change
    const columns = React.useMemo(() => generateColumns(columnStyles), [
        columnStyles
    ]);
    const tableContext = React.useMemo(
        () => ({ columns, state: { selectedExecution, setSelectedExecution } }),
        [columns, selectedExecution, setSelectedExecution]
    );

    const onCloseDetailsPanel = () => setSelectedExecution(null);

    const rowProps = {
        selectedExecution,
        setSelectedExecution
    };
    const content =
        executionsWithKeys.length > 0 ? (
            executionsWithKeys.map(({ nodeExecution, cacheKey }, index) => {
                return (
                    <NodeExecutionRow
                        {...rowProps}
                        abortMetadata={abortMetadata}
                        index={index}
                        key={cacheKey}
                        execution={nodeExecution}
                    />
                );
            })
        ) : (
            <NoExecutionsContent size="large" />
        );

    return (
        <div
            className={classnames(
                tableStyles.tableContainer,
                commonStyles.flexFill
            )}
        >
            <ExecutionsTableHeader
                columns={columns}
                scrollbarPadding={scrollbarPadding}
            />
            <NodeExecutionsTableContext.Provider value={tableContext}>
                <div className={tableStyles.scrollContainer}>{content}</div>
            </NodeExecutionsTableContext.Provider>
            <DetailsPanel
                open={selectedExecution !== null}
                onClose={onCloseDetailsPanel}
            >
                {selectedExecution != null ? (
                    <NodeExecutionDetailsPanelContent
                        onClose={onCloseDetailsPanel}
                        nodeExecutionId={selectedExecution}
                    />
                ) : null}
            </DetailsPanel>
        </div>
    );
};
