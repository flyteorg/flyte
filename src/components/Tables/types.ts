import { TableCellProps as BaseTableCellProps } from 'react-virtualized';
export type LoadMoreRowItem = symbol;
/** A fixed row height shared across all rows, or a callback which receives an item index and computes the height */
export type RowHeight = number | ((index: number) => number);

export interface TableCellProps<T> extends BaseTableCellProps {
    cellData?: T;
}
