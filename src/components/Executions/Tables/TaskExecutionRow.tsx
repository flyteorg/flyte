import * as classnames from 'classnames';
import { useExpandableMonospaceTextStyles } from 'components/common/ExpandableMonospaceText';
import { TaskExecution } from 'models';
import * as React from 'react';
import { DetailedNodeExecution, DetailedTaskExecution } from '../types';
import { NodeExecutionsTableContext } from './contexts';
import { ExpandableExecutionError } from './ExpandableExecutionError';
import { RowExpander } from './RowExpander';
import { useExecutionTableStyles } from './styles';
import { TaskExecutionChildren } from './TaskExecutionChildren';
import { TaskExecutionColumnDefinition } from './types';

interface TaskExecutionRowProps {
    columns: TaskExecutionColumnDefinition[];
    execution: DetailedTaskExecution;
    nodeExecution: DetailedNodeExecution;
    style?: React.CSSProperties;
}

function isExpandableExecution(execution: TaskExecution) {
    return execution.isParent;
}

/** Renders a TaskExecution as a row inside a `NodeExecutionsTable` */
export const TaskExecutionRow: React.FC<TaskExecutionRowProps> = ({
    columns,
    execution,
    nodeExecution,
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

    const errorContent = null;
    // TODO: Show errors or remove this
    // const errorContent = error ? (
    //     <ExpandableExecutionError
    //         onExpandCollapse={onHeightChange}
    //         error={error}
    //     />
    // ) : null;

    const extraContent = expanded ? (
        <div
            className={classnames(
                tableStyles.childrenContainer,
                monospaceTextStyles.nestedParent
            )}
        >
            {errorContent}
            <TaskExecutionChildren taskExecution={execution} />
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
                            nodeExecution,
                            state
                        })}
                    </div>
                ))}
            </div>
            {extraContent}
        </div>
    );
};
