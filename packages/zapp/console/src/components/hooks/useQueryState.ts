import { pickBy } from 'lodash';
import { parse, ParsedQuery, stringify } from 'query-string';
import { useEffect, useState } from 'react';
import { history } from 'routes/history';

/** A hook which allows reading/setting of the current query string params
 * It will attach a listener to history so that components using query params
 * are updated whenever the path/query changes.
 */
export const useQueryState = <T extends ParsedQuery>() => {
  const [params, setParams] = useState<Partial<T>>(parse(history.location.search) as Partial<T>);

  const setQueryState = (newState: T) => {
    // Remove any undefined values before serializing
    const finalState = pickBy(newState, (v) => v !== undefined);
    history.replace({ ...history.location, search: stringify(finalState) });
  };

  const setQueryStateValue = (key: string, value?: string) => {
    const newParams = { ...params, [key]: value } as T;
    setQueryState(newParams);
  };

  useEffect(() => {
    return history.listen((location) => {
      setParams(parse(location.search) as Partial<T>);
    });
  }, [history]);

  return {
    params,
    setQueryState,
    setQueryStateValue,
  };
};
