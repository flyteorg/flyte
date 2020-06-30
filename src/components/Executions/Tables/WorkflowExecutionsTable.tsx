import { Typography } from '@material-ui/core';
import Link from '@material-ui/core/Link';
import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { noExecutionsFoundString } from 'common/constants';
import {
    dateFromNow,
    formatDateLocalTimezone,
    formatDateUTC,
    millisecondsToHMS
} from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { DataList, DataListRef } from 'components';
import { ListProps } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { Execution } from 'models';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ListRowRenderer } from 'react-virtualized';
import { ExecutionInputsOutputsModal } from '../ExecutionInputsOutputsModal';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { getWorkflowExecutionTimingMS } from '../utils';
import { workflowExecutionsTableColumnWidths } from './constants';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { ExpandableExecutionError } from './ExpandableExecutionError';
import { useExecutionTableStyles } from './styles';
import { ColumnDefinition } from './types';
import { WorkflowExecutionLink } from './WorkflowExecutionLink';

const useStyles = makeStyles((theme: Theme) => ({
    cellName: {
        paddingLeft: theme.spacing(1)
    },
    columnName: {
        flexBasis: workflowExecutionsTableColumnWidths.name
    },
    columnLastRun: {
        flexBasis: workflowExecutionsTableColumnWidths.lastRun
    },
    columnStatus: {
        flexBasis: workflowExecutionsTableColumnWidths.phase
    },
    columnStartedAt: {
        flexBasis: workflowExecutionsTableColumnWidths.startedAt
    },
    columnDuration: {
        flexBasis: workflowExecutionsTableColumnWidths.duration,
        textAlign: 'right'
    },
    columnInputsOutputs: {
        flexGrow: 1,
        flexBasis: workflowExecutionsTableColumnWidths.inputsOutputs,
        marginLeft: theme.spacing(2),
        marginRight: theme.spacing(2),
        textAlign: 'left'
    },
    row: {
        paddingLeft: theme.spacing(2)
    }
}));

function useWorkflowExecutionsTableState() {
    const [
        selectedIOExecution,
        setSelectedIOExecution
    ] = React.useState<Execution | null>(null);
    return {
        selectedIOExecution,
        setSelectedIOExecution
    };
}

interface WorkflowExecutionCellRendererData {
    execution: Execution;
    state: ReturnType<typeof useWorkflowExecutionsTableState>;
}
type WorkflowExecutionColumnDefinition = ColumnDefinition<
    WorkflowExecutionCellRendererData
>;

function generateColumns(
    styles: ReturnType<typeof useStyles>,
    tableStyles: ReturnType<typeof useExecutionTableStyles>
): WorkflowExecutionColumnDefinition[] {
    return [
        {
            cellRenderer: ({
                execution: {
                    id,
                    closure: { startedAt }
                }
            }) => (
                <>
                    <WorkflowExecutionLink id={id} />
                    <Typography variant="subtitle1" color="textSecondary">
                        {startedAt
                            ? `Last run ${dateFromNow(
                                  timestampToDate(startedAt)
                              )}`
                            : ''}
                    </Typography>
                </>
            ),
            className: styles.columnName,
            key: 'name',
            label: 'execution id'
        },
        {
            cellRenderer: ({
                execution: {
                    closure: { phase = WorkflowExecutionPhase.UNDEFINED }
                }
            }) => <ExecutionStatusBadge phase={phase} type="workflow" />,
            className: styles.columnStatus,
            key: 'phase',
            label: 'status'
        },
        {
            cellRenderer: ({ execution: { closure } }) => {
                const { startedAt } = closure;
                if (!startedAt) {
                    return '';
                }
                const startedAtDate = timestampToDate(startedAt);
                return (
                    <>
                        <Typography variant="body1">
                            {formatDateUTC(startedAtDate)}
                        </Typography>
                        <Typography variant="subtitle1" color="textSecondary">
                            {formatDateLocalTimezone(startedAtDate)}
                        </Typography>
                    </>
                );
            },
            className: styles.columnStartedAt,
            key: 'startedAt',
            label: 'start time'
        },
        {
            cellRenderer: ({ execution }) => {
                const timing = getWorkflowExecutionTimingMS(execution);
                return (
                    <Typography variant="body1">
                        {timing !== null
                            ? millisecondsToHMS(timing.duration)
                            : ''}
                    </Typography>
                );
            },
            className: styles.columnDuration,
            key: 'duration',
            label: 'duration'
        },
        {
            cellRenderer: ({ execution, state }) => {
                const onClick = () => state.setSelectedIOExecution(execution);
                return (
                    <Link component="button" onClick={onClick} variant="body1">
                        View Inputs &amp; Outputs
                    </Link>
                );
            },
            className: styles.columnInputsOutputs,
            key: 'inputsOutputs',
            label: ''
        }
    ];
}

export interface WorkflowExecutionsTableProps extends ListProps<Execution> {}

/** Renders a table of WorkflowExecution records. Executions with errors will
 * have an expanadable container rendered as part of the table row.
 */
export const WorkflowExecutionsTable: React.FC<WorkflowExecutionsTableProps> = props => {
    const executions = props.value;
    const state = useWorkflowExecutionsTableState();
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    const tableStyles = useExecutionTableStyles();
    const listRef = React.useRef<DataListRef>(null);
    // Memoizing columns so they won't be re-generated unless the styles change
    const columns = React.useMemo(() => generateColumns(styles, tableStyles), [
        styles,
        commonStyles
    ]);

    const retry = () => props.fetch();
    const onCloseIOModal = () => state.setSelectedIOExecution(null);
    const recomputeRow = (rowIndex: number) => {
        if (listRef.current !== null) {
            listRef.current.recomputeRowHeights(rowIndex);
        }
    };

    // Custom renderer to allow us to append error content to executions which
    // are in a failed state
    const rowRenderer: ListRowRenderer = rowProps => {
        const { index: rowIndex, style } = rowProps;
        const execution = executions[rowIndex];
        const { error } = execution.closure;

        return (
            <div
                className={classnames(
                    tableStyles.row,
                    styles.row,
                    tableStyles.borderBottom
                )}
                style={style}
            >
                <div className={tableStyles.rowColumns}>
                    {columns.map(
                        ({ className, key: columnKey, cellRenderer }) => (
                            <div
                                key={columnKey}
                                className={classnames(
                                    tableStyles.rowColumn,
                                    className
                                )}
                            >
                                {cellRenderer({
                                    execution,
                                    state
                                })}
                            </div>
                        )
                    )}
                </div>
                {error ? (
                    <ExpandableExecutionError
                        onExpandCollapse={recomputeRow.bind(null, rowIndex)}
                        error={error}
                    />
                ) : null}
            </div>
        );
    };

    return (
        <div
            className={classnames(
                tableStyles.tableContainer,
                commonStyles.flexFill
            )}
        >
            <ExecutionsTableHeader columns={columns} />
            <DataList
                {...props}
                onRetry={retry}
                noRowsContent={noExecutionsFoundString}
                ref={listRef}
                rowContentRenderer={rowRenderer}
            />
            <ExecutionInputsOutputsModal
                execution={state.selectedIOExecution}
                onClose={onCloseIOModal}
            />
        </div>
    );
};
