import { FetchableState } from 'components/hooks/types';
import { ScrollbarPresenceParams } from 'react-virtualized';

export interface ListProps<T> {
    // height/width are optional. If unspecified, the component will
    // use auto-sizing behavior
    height?: number;
    value: T[];
    lastError: string | Error | null;
    state: FetchableState<T[]>;
    moreItemsAvailable: boolean;
    onScrollbarPresenceChange?: (params: ScrollbarPresenceParams) => any;
    width?: number;
    fetch(): void;
}
