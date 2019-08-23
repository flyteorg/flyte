import * as classnames from 'classnames';
import { useExpandableMonospaceTextStyles } from 'components/common/ExpandableMonospaceText';
import * as React from 'react';
import { DetailedNodeExecution } from '../types';
import { NodeExecutionsTableContext } from './contexts';
import { ExpandableExecutionError } from './ExpandableExecutionError';
import { NodeExecutionChildren } from './NodeExecutionChildren';
import { RowExpander } from './RowExpander';
import { selectedClassName, useExecutionTableStyles } from './styles';
import { NodeExecutionColumnDefinition } from './types';

interface NodeExecutionRowProps {
    columns: NodeExecutionColumnDefinition[];
    execution: DetailedNodeExecution;
    style?: React.CSSProperties;
}

function isExpandableExecution(execution: DetailedNodeExecution) {
    return true;
}

/** Renders a NodeExecution as a row inside a `NodeExecutionsTable` */
export const NodeExecutionRow: React.FC<NodeExecutionRowProps> = ({
    columns,
    execution,
    style
}) => {
    const state = React.useContext(NodeExecutionsTableContext);
    const [expanded, setExpanded] = React.useState(false);
    const toggleExpanded = () => {
        setExpanded(!expanded);
    };
    const tableStyles = useExecutionTableStyles();
    const monospaceTextStyles = useExpandableMonospaceTextStyles();

    const selected = state.selectedExecution
        ? state.selectedExecution === execution
        : false;
    const { error } = execution.closure;

    const expanderContent = isExpandableExecution(execution) ? (
        <RowExpander expanded={expanded} onClick={toggleExpanded} />
    ) : null;

    const errorContent = error ? (
        <ExpandableExecutionError error={error} />
    ) : null;

    const extraContent = expanded ? (
        <div
            className={classnames(
                tableStyles.childrenContainer,
                monospaceTextStyles.nestedParent
            )}
        >
            {errorContent}
            <NodeExecutionChildren execution={execution} />
        </div>
    ) : (
        errorContent
    );

    return (
        <div
            className={classnames(tableStyles.row, {
                [selectedClassName]: selected
            })}
            style={style}
        >
            <div className={tableStyles.rowColumns}>
                <div className={tableStyles.expander}>{expanderContent}</div>
                {columns.map(({ className, key: columnKey, cellRenderer }) => (
                    <div
                        key={columnKey}
                        className={classnames(tableStyles.rowColumn, className)}
                    >
                        {cellRenderer({
                            execution,
                            state
                        })}
                    </div>
                ))}
            </div>
            {extraContent}
        </div>
    );
};
