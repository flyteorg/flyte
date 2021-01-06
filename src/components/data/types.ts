import { QueryObserverOptions } from 'react-query';
import { QueryType } from './queries';

type QueryKeyArray = [QueryType, ...unknown[]];
export interface QueryInput<T> extends QueryObserverOptions<T, Error> {
    queryKey: QueryKeyArray;
    queryFn: QueryObserverOptions<T, Error>['queryFn'];
}
