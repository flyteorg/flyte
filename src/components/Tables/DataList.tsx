import { ListProps, Typography } from '@material-ui/core';
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import { tablePlaceholderColor } from 'components/Theme/constants';
import * as React from 'react';
import {
    AutoSizer,
    CellMeasurer,
    CellMeasurerCache,
    Index,
    List,
    ListRowRenderer
} from 'react-virtualized';
import { headerGridHeight, tableGridPadding } from './constants';
import { LoadMoreRowContent } from './LoadMoreRowContent';

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
    noRowsContent: {
        color: tablePlaceholderColor,
        margin: `${theme.spacing(5)}px auto`,
        textAlign: 'center'
    }
}));

export interface DataListRowClickEventParams<T> {
    index: number;
    data: T;
}

export interface DataListProps extends ListProps<any> {
    className?: string;
    height?: number;
    noRowsContent?: string | React.ComponentType;
    width?: number;
    ref?: React.Ref<DataListRef>;
    rowContentRenderer: ListRowRenderer;
    onRetry?(): void;
    onRowClick?(params: DataListRowClickEventParams<any>): void;
}

// When using DataListImpl directly, the auto-sizing behavior is removed.
// So width/height are explicitly required.
export interface DataListImplProps extends DataListProps {
    height: number;
    width: number;
}

const createCellMeasurerCache = () =>
    new CellMeasurerCache({
        fixedWidth: true
    });

export interface DataListRef {
    recomputeRowHeights(index: number): void;
}

const DataListImplComponent: React.RefForwardingComponent<
    DataListRef,
    DataListImplProps
> = (props, ref) => {
    const {
        height,
        value: items,
        lastError,
        isFetching,
        moreItemsAvailable,
        onScrollbarPresenceChange,
        noRowsContent: NoRowsContent = 'No items found.',
        rowContentRenderer,
        width
    } = props;

    const styles = useStyles();
    const theme = useTheme<Theme>();
    const lengthRef = React.useRef<number>(0);
    const listRef = React.useRef<List>(null);
    /** We want the cache to persist across renders, which useState will do.
     * But we also don't want to be needlessly creating new caches that are
     * thrown away immediately. So we're using a creation function which useState
     * will call to create the initial value.
     */
    const [cellCache] = React.useState(createCellMeasurerCache);

    const recomputeRow = React.useMemo(
        () => (rowIndex: number) => {
            cellCache.clear(rowIndex, 0);
            if (listRef.current !== null) {
                listRef.current.recomputeRowHeights(rowIndex);
            }
        },
        [cellCache, listRef]
    );
    React.useImperativeHandle(
        ref,
        () => ({
            recomputeRowHeights: recomputeRow
        }),
        [recomputeRow]
    );

    React.useLayoutEffect(() => {
        if (lengthRef.current >= 0 && items.length > lengthRef.current) {
            recomputeRow(lengthRef.current);
        }
        lengthRef.current = props.value.length;
    }, [items.length]);

    const headerHeight = theme.spacing(headerGridHeight);
    const listPadding = theme.spacing(tableGridPadding);
    const showLoadMore = items.length && moreItemsAvailable;

    // We want the "load more" content to render at the end of the list.
    // We want the list to fill all available space and only have one scrollbar.
    // for all content.
    // So we make the list the only content which
    // can scroll, let react-virtualized treat the "load more" like the
    // last row of the list, and then render special content for the last index.
    const rowCount = showLoadMore ? items.length + 1 : items.length;

    const noRowsRenderer = () => (
        <div className={styles.noRowsContent}>
            {typeof NoRowsContent === 'string' ? (
                <Typography variant="h6">{NoRowsContent}</Typography>
            ) : (
                <NoRowsContent />
            )}
        </div>
    );

    const rowRenderer: ListRowRenderer = rowProps => {
        const { key, index, parent, style } = rowProps;
        let content: React.ReactNode;
        if (index === items.length) {
            content = (
                <LoadMoreRowContent
                    loadMoreRows={loadMoreRows}
                    isFetching={isFetching}
                    lastError={lastError}
                    style={style}
                />
            );
        } else {
            content = rowContentRenderer(rowProps);
        }

        return (
            <CellMeasurer
                cache={cellCache}
                columnIndex={0}
                key={key}
                rowIndex={index}
                parent={parent}
            >
                {content}
            </CellMeasurer>
        );
    };

    const rowGetter = ({ index }: Index) => {
        // The last item trick will mean we try to read past the end of the
        // list. We don't end up rendering this rowData, but we still need it
        // to be valid.
        const clamped = Math.min(index, items.length - 1);
        return items[clamped];
    };

    const loadMoreRows = () => props.fetch();

    return (
        <>
            <List
                deferredMeasurementCache={cellCache}
                headerHeight={headerHeight}
                headerClassName={styles.headerItem}
                height={height - listPadding * 2}
                noRowsRenderer={noRowsRenderer}
                onScrollbarPresenceChange={onScrollbarPresenceChange}
                ref={listRef}
                rowCount={rowCount}
                rowGetter={rowGetter}
                rowRenderer={rowRenderer}
                rowHeight={cellCache.rowHeight}
                width={width}
            />
        </>
    );
};
const DataListImpl = React.forwardRef(DataListImplComponent);

/** The default version of DataList doesn't require a width/height and will expand to
 * fill its parent container (this can have odd behavior when using flex or the parent
 * doesn't have an explicit width/height set). When passing an explicit width/height,
 * the auto-sizing behavior is disabled and a slightly simpler component tree will be rendered.
 */
const DataListComponent: React.RefForwardingComponent<
    DataListRef,
    DataListProps
> = (props, ref) => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();

    if (props.width && props.height) {
        return <DataListImpl {...(props as DataListImplProps)} />;
    }
    return (
        <div
            className={classnames(styles.listContainer, commonStyles.flexFill)}
        >
            <div className={commonStyles.flexFill}>
                <AutoSizer>
                    {({ height, width }) => (
                        <DataListImpl
                            {...props}
                            ref={ref}
                            width={width}
                            height={height}
                        />
                    )}
                </AutoSizer>
            </div>
        </div>
    );
};
const DataList = React.forwardRef(DataListComponent);

export { DataList, DataListImpl };
