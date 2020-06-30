import * as classnames from 'classnames';
import { DetailsPanel, ListProps } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import * as scrollbarSize from 'dom-helpers/util/scrollbarSize';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { NodeExecutionDetails } from '../ExecutionDetails/NodeExecutionDetails';
import { NodeExecutionsTableContext } from './contexts';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { generateColumns } from './nodeExecutionColumns';
import { NodeExecutionRow } from './NodeExecutionRow';
import { NoExecutionsContent } from './NoExecutionsContent';
import { useColumnStyles, useExecutionTableStyles } from './styles';
import { useNodeExecutionsTableState } from './useNodeExecutionsTableState';

export interface NodeExecutionsTableProps extends ListProps<NodeExecution> {}

const scrollbarPadding = scrollbarSize();

/** Renders a table of NodeExecution records. Executions with errors will
 * have an expanadable container rendered as part of the table row.
 * NodeExecutions are expandable and will potentially render a list of child
 * TaskExecutions
 */
export const NodeExecutionsTable: React.FC<NodeExecutionsTableProps> = props => {
    const state = useNodeExecutionsTableState(props);
    const commonStyles = useCommonStyles();
    const tableStyles = useExecutionTableStyles();

    const columnStyles = useColumnStyles();
    // Memoizing columns so they won't be re-generated unless the styles change
    const columns = React.useMemo(() => generateColumns(columnStyles), [
        columnStyles
    ]);

    const onCloseDetailsPanel = () => state.setSelectedExecution(null);

    const rowProps = { state, onHeightChange: () => {} };
    const content =
        state.executions.length > 0 ? (
            state.executions.map((execution, index) => {
                return (
                    <NodeExecutionRow
                        {...rowProps}
                        index={index}
                        key={execution.cacheKey}
                        execution={execution}
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
            <NodeExecutionsTableContext.Provider value={{ columns, state }}>
                <div className={tableStyles.scrollContainer}>{content}</div>
            </NodeExecutionsTableContext.Provider>
            <DetailsPanel
                open={state.selectedExecution !== null}
                onClose={onCloseDetailsPanel}
            >
                {state.selectedExecution && (
                    <NodeExecutionDetails
                        onClose={onCloseDetailsPanel}
                        execution={state.selectedExecution}
                    />
                )}
            </DetailsPanel>
        </div>
    );
};
