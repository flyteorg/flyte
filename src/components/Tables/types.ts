import { TableCellProps } from 'react-virtualized';
export type LoadMoreRowItem = Symbol;
/** A fixed row height shared across all rows, or a callback which receives an item index and computes the height */
export type RowHeight = number | ((index: number) => number);

export interface TableCellProps<T> extends TableCellProps {
    cellData?: T;
}
