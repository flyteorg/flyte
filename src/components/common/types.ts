import { ScrollbarPresenceParams } from 'react-virtualized';

export interface ListProps<T> {
    // height/width are optional. If unspecified, the component will
    // use auto-sizing behavior
    height?: number;
    value: T[];
    lastError: string | Error | null;
    loading: boolean;
    moreItemsAvailable: boolean;
    onScrollbarPresenceChange?: (params: ScrollbarPresenceParams) => any;
    width?: number;
    fetch(): Promise<unknown>;
}
