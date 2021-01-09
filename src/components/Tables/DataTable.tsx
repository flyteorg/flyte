import { Typography } from '@material-ui/core';
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { noExecutionsFoundString } from 'common/constants';
import { ListProps } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import {
    bodyFontSize,
    headerFontFamily,
    separatorColor,
    tableHeaderColor,
    tablePlaceholderColor
} from 'components/Theme';
import * as React from 'react';
import {
    AutoSizer,
    Column,
    ColumnProps,
    defaultTableRowRenderer,
    Index,
    Table,
    TableRowRenderer
} from 'react-virtualized';
import {
    headerGridHeight,
    loadMoreRowGridHeight,
    tableGridPadding
} from './constants';
import { LoadMoreRowContent } from './LoadMoreRowContent';
import { RowHeight } from './types';

const useStyles = makeStyles((theme: Theme) => ({
    headerItem: {
        alignItems: 'center',
        display: 'flex',
        height: '100%'
    },
    listContainer: {
        display: 'flex',
        flexDirection: 'column'
    },
    loadMoreContainer: {
        bottom: 0,
        height: theme.spacing(loadMoreRowGridHeight),
        left: 0,
        position: 'absolute',
        right: 0
    },
    noRowsContent: {
        color: tablePlaceholderColor,
        margin: `${theme.spacing(5)}px auto`,
        textAlign: 'center'
    },
    table: {
        position: 'relative',
        marginBottom: theme.spacing(loadMoreRowGridHeight),
        '& .ReactVirtualized__Table__headerRow': {
            alignItems: 'center',
            borderBottom: `4px solid ${separatorColor}`,
            borderTop: `1px solid ${separatorColor}`,
            color: tableHeaderColor,
            display: 'flex',
            fontFamily: headerFontFamily,
            fontSize: bodyFontSize,
            fontWeight: 'bold',
            flexDirection: 'row',
            textTransform: 'uppercase'
        },
        '& .ReactVirtualized__Table__row': {
            alignItems: 'center',
            borderBottom: `1px solid ${separatorColor}`,
            display: 'flex',
            flexDirection: 'row'
        },
        '& .ReactVirtualized__Table__headerColumn': {
            marginRight: theme.spacing(1),
            minWidth: 0,
            '&:first-of-type': {
                marginLeft: theme.spacing(2)
            }
        },
        '& .ReactVirtualized__Table__rowColumn': {
            marginRight: theme.spacing(1),
            minWidth: 0,
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            '&:first-of-type': {
                marginLeft: theme.spacing(2),
                paddingLeft: theme.spacing(1)
            }
        }
    }
}));

export type KeyedColumnProps = ColumnProps & { key: string };
export interface DataTableRowClickEventParams<T> {
    index: number;
    data: T;
}

export interface DataTableProps extends ListProps<any> {
    columns: KeyedColumnProps[];
    className?: string;
    height?: number;
    rowHeight: RowHeight;
    width?: number;
    rowContentRenderer?: TableRowRenderer;
    onRetry?(): void;
    onRowClick(params: DataTableRowClickEventParams<any>): void;
}

// When using DataTableImpl directly, the auto-sizing behavior is removed.
// So width/height are explicitly required.
export interface DataTableImplProps extends DataTableProps {
    height: number;
    width: number;
}

export const DataTableImpl: React.FC<DataTableImplProps> = props => {
    const styles = useStyles();
    const theme = useTheme<Theme>();
    const headerHeight = theme.spacing(headerGridHeight);
    const tablePadding = theme.spacing(tableGridPadding);
    const loadMoreRowHeight = theme.spacing(loadMoreRowGridHeight);
    const {
        columns,
        height,
        value: items,
        lastError,
        isFetching,
        moreItemsAvailable,
        rowContentRenderer = defaultTableRowRenderer,
        width
    } = props;
    const showLoadMore = items.length && moreItemsAvailable;
    const rowCount = showLoadMore ? items.length + 1 : items.length;

    const noRowsRenderer = () => (
        <div className={styles.noRowsContent}>
            <Typography variant="h6">{noExecutionsFoundString}</Typography>
        </div>
    );

    const rowRenderer: TableRowRenderer = rowProps => {
        const { className, index, style } = rowProps;
        if (index === items.length) {
            return (
                <div className={styles.loadMoreContainer}>
                    <LoadMoreRowContent
                        className={className}
                        loadMoreRows={loadMoreRows}
                        isFetching={isFetching}
                        lastError={lastError}
                        style={style}
                    />
                </div>
            );
        }

        return rowContentRenderer(rowProps);
    };

    const rowGetter = ({ index }: Index) => {
        // The last item trick will mean we try to read past the end of the
        // list. We don't end up rendering this rowData, but we still need it
        // to be valid so that any custom cellDataGetters are safe to access
        // properties on it.
        const clamped = Math.min(index, items.length - 1);
        return items[clamped];
    };

    const loadMoreRows = () => props.fetch();

    const getRowHeight: (info: Index) => number = ({ index }) => {
        if (index === items.length) {
            return loadMoreRowHeight;
        }

        return typeof props.rowHeight === 'function'
            ? props.rowHeight(index)
            : props.rowHeight;
    };

    return (
        <>
            <Table
                className={styles.table}
                headerHeight={headerHeight}
                headerClassName={styles.headerItem}
                height={height - tablePadding * 2}
                noRowsRenderer={noRowsRenderer}
                rowCount={rowCount}
                rowGetter={rowGetter}
                rowRenderer={rowRenderer}
                rowHeight={getRowHeight}
                width={width - tablePadding * 2}
            >
                {columns.map(props => (
                    <Column {...props} key={props.key} />
                ))}
            </Table>
        </>
    );
};

/** The default version of DataTable doesn't require a width/height and will expand to
 * fill its parent container (this can have odd behavior when using flex or the parent
 * doesn't have an explicit width/height set). When passing an explicit width/height,
 * the auto-sizing behavior is disabled and a slightly simpler component tree will be rendered.
 */
export const DataTable = (props: DataTableProps) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();

    if (props.width && props.height) {
        return <DataTableImpl {...(props as DataTableImplProps)} />;
    }
    return (
        <div
            className={classnames(styles.listContainer, commonStyles.flexFill)}
        >
            <div className={commonStyles.flexFill}>
                <AutoSizer>
                    {({ height, width }) => (
                        <DataTableImpl
                            {...props}
                            width={width}
                            height={height}
                        />
                    )}
                </AutoSizer>
            </div>
        </div>
    );
};
