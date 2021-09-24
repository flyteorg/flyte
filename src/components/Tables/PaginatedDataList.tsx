import * as React from 'react';
import classnames from 'classnames';
import { createStyles, makeStyles, Theme } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
// import TablePagination from '@material-ui/core/TablePagination';
import TableRow from '@material-ui/core/TableRow';
import TableSortLabel from '@material-ui/core/TableSortLabel';
import Paper from '@material-ui/core/Paper';
import { ColumnDefinition } from '../Executions/Tables/types';
import { PropsWithChildren } from 'react';
import { Typography } from '@material-ui/core';
import { headerFontFamily, tableHeaderColor } from '../Theme/constants';
import { headerGridHeight } from './constants';
import { workflowVersionsTableColumnWidths } from '../Executions/Tables/constants';

type Order = 'asc' | 'desc';

interface PaginatedDataListHeaderProps<C> {
    classes: ReturnType<typeof useStyles>;
    columns: ColumnDefinition<any>[];
    onRequestSort: (event: React.MouseEvent<unknown>, property: string) => void;
    order: Order;
    orderBy: string;
    rowCount: number;
    showRadioButton?: boolean;
}

interface PaginatedDataListProps<T> {
    columns: ColumnDefinition<any>[];
    showRadioButton?: boolean;
    totalRows: number;
    data: T[];
    rowRenderer: (row: T) => void;
    noDataString: string;
    fillEmptyRows?: boolean;
}

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        root: {
            width: '100%'
        },
        paper: {
            width: '100%',
            marginBottom: theme.spacing(2)
        },
        table: {
            minWidth: 750
        },
        radioButton: {
            width: workflowVersionsTableColumnWidths.radio
        },
        header: {
            borderBottom: `4px solid ${theme.palette.divider}`,
            borderTop: `1px solid ${theme.palette.divider}`,
            color: tableHeaderColor,
            fontFamily: headerFontFamily,
            height: theme.spacing(headerGridHeight)
        },
        cell: {
            padding: theme.spacing(1),
            textTransform: 'uppercase'
        },
        visuallyHidden: {
            border: 0,
            clip: 'rect(0 0 0 0)',
            height: 1,
            margin: -1,
            overflow: 'hidden',
            padding: 0,
            position: 'absolute',
            top: 20,
            width: 1
        }
    })
);

/**
 * Renders pagination table's header
 * @param props
 * @constructor
 */
const PaginatedDataListHeader = <C,>(
    props: PropsWithChildren<PaginatedDataListHeaderProps<C>>
) => {
    const {
        classes,
        order,
        orderBy,
        onRequestSort,
        columns,
        showRadioButton
    } = props;
    const createSortHandler = (property: string) => (
        event: React.MouseEvent<unknown>
    ) => {
        onRequestSort(event, property);
    };
    const cellClass = classnames(classes.cell, classes.header);

    return (
        <TableHead>
            <TableRow>
                {showRadioButton && (
                    <TableCell
                        classes={{
                            root: cellClass
                        }}
                        className={classes.radioButton}
                    >
                        &nbsp;
                    </TableCell>
                )}
                {columns.map(column => (
                    <TableCell
                        classes={{
                            root: cellClass
                        }}
                        className={column.className}
                        key={column.key}
                        align="left"
                        padding="default"
                        sortDirection={orderBy === column.key ? order : false}
                    >
                        <TableSortLabel
                            active={orderBy === column.key}
                            direction={orderBy === column.key ? order : 'asc'}
                            onClick={createSortHandler(column.key)}
                        >
                            <Typography variant="overline">
                                {column.label}
                            </Typography>
                            {orderBy === column.key ? (
                                <span className={classes.visuallyHidden}>
                                    {order === 'desc'
                                        ? 'sorted descending'
                                        : 'sorted ascending'}
                                </span>
                            ) : null}
                        </TableSortLabel>
                    </TableCell>
                ))}
            </TableRow>
        </TableHead>
    );
};

/**
 * Renders pagination table based on the column information and the total number of rows.
 * @param columns
 * @param data
 * @param rowRenderer
 * @param totalRows
 * @param showRadioButton
 * @param noDataString
 * @constructor
 */
const PaginatedDataList = <T,>({
    columns,
    data,
    rowRenderer,
    totalRows,
    showRadioButton,
    fillEmptyRows = true,
    noDataString
}: PropsWithChildren<PaginatedDataListProps<T>>) => {
    const classes = useStyles();
    const [order, setOrder] = React.useState<Order>('asc');
    const [orderBy, setOrderBy] = React.useState('calories');
    const [page, setPage] = React.useState(0);
    const [rowsPerPage, setRowsPerPage] = React.useState(5);

    const handleRequestSort = (
        event: React.MouseEvent<unknown>,
        property: string
    ) => {
        const isAsc = orderBy === property && order === 'asc';
        setOrder(isAsc ? 'desc' : 'asc');
        setOrderBy(property);
    };

    // const handleChangePage = (event: unknown, newPage: number) => {
    //     setPage(newPage);
    // };
    //
    // const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    //     setRowsPerPage(parseInt(event.target.value, 10));
    //     setPage(0);
    // };

    const emptyRows =
        rowsPerPage - Math.min(rowsPerPage, totalRows - page * rowsPerPage);

    return (
        <div className={classes.root}>
            <Paper className={classes.paper}>
                <TableContainer>
                    <Table className={classes.table} size="medium">
                        <PaginatedDataListHeader
                            classes={classes}
                            order={order}
                            orderBy={orderBy}
                            onRequestSort={handleRequestSort}
                            columns={columns}
                            rowCount={data.length}
                            showRadioButton={showRadioButton}
                        />
                        <TableBody>
                            {data.map((row, index) => {
                                return rowRenderer(row);
                            })}
                            {fillEmptyRows &&
                                !showRadioButton &&
                                emptyRows > 0 && (
                                    <TableRow
                                        style={{
                                            height: `${4 * emptyRows}rem`
                                        }}
                                    >
                                        <TableCell colSpan={6} />
                                    </TableRow>
                                )}
                        </TableBody>
                    </Table>
                </TableContainer>
                {/*<TablePagination*/}
                {/*    rowsPerPageOptions={[5, 10, 25]}*/}
                {/*    component="div"*/}
                {/*    count={totalRows}*/}
                {/*    rowsPerPage={rowsPerPage}*/}
                {/*    page={page}*/}
                {/*    onChangePage={handleChangePage}*/}
                {/*    onChangeRowsPerPage={handleChangeRowsPerPage}*/}
                {/*/>*/}
            </Paper>
        </div>
    );
};

export default PaginatedDataList;
