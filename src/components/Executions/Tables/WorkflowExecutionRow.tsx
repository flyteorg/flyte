import * as classnames from 'classnames';
import { useExpandableMonospaceTextStyles } from 'components/common/ExpandableMonospaceText';
import { Execution } from 'models/Execution';
import * as React from 'react';
import { NodeExecutionsTableContext } from './contexts';
import { ExpandableExecutionError } from './ExpandableExecutionError';
import { RowExpander } from './RowExpander';
import { useExecutionTableStyles } from './styles';
import { WorkflowExecutionColumnDefinition } from './types';
import { WorkflowExecutionChildren } from './WorkflowExecutionChildren';
// import { WorkflowExecutionChildren } from './WorkflowExecutionChildren';

interface WorkflowExecutionRowProps {
    columns: WorkflowExecutionColumnDefinition[];
    execution: Execution;
    style?: React.CSSProperties;
}

function isExpandableExecution(execution: Execution) {
    return true;
}

/** Renders a WorkflowExecution as a row inside a `NodeExecutionsTable` */
export const WorkflowExecutionRow: React.FC<WorkflowExecutionRowProps> = ({
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
            <WorkflowExecutionChildren execution={execution} />
        </div>
    ) : (
        errorContent
    );

    return (
        <div className={tableStyles.row} style={style}>
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
