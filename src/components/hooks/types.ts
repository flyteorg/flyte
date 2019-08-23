import {
    Execution,
    NodeExecution,
    PaginatedEntityResponse,
    RequestConfig
} from 'models';

export interface FetchArgs {
    force?: boolean;
}

export interface FetchableData<T> {
    fetch(fetchArgs?: FetchArgs): Promise<T>;
    hasLoaded: boolean;
    lastError: Error | null;
    loading: boolean;
    debugName: string;
    value: T;
}

export interface FetchableExecution {
    fetchable: FetchableData<Execution>;
    terminateExecution(cause: string): Promise<void>;
}

export interface PaginatedFetchableData<T> extends FetchableData<T[]> {
    /** Whether or not a fetch would yield more items. Useful for determining if
     * a "load more" button should be shown
     */
    moreItemsAvailable: boolean;
}

export type FetchFn<T, DataType> = (
    data: DataType,
    currentValue?: T
) => Promise<T>;

export type PaginatedFetchFn<T, DataType = undefined> = (
    data: DataType,
    config: RequestConfig
) => Promise<PaginatedEntityResponse<T>>;

/** Defines parameters for auto-refreshing an entity of type T */
export interface RefreshConfig<T> {
    /** How often (ms) to check the finality of the entity and perform a refresh */
    interval: number;
    /** Implements a boolean check for whether this entity is in a final state
     * and should no longer be refreshed
     */
    valueIsFinal: (value: T) => boolean;
    /** Implements the fetch logic for this entity. Defaults to calling fetch() */
    doRefresh?: (fetchable: FetchableData<T>) => Promise<T>;
}

export type NodeExecutionDictionary = Record<string, NodeExecution>;
