import * as fuzzysort from 'fuzzysort';
import { createElement, Fragment, useEffect, useState } from 'react';

interface SearchTarget<T> {
  value: T;
  prepared: Fuzzysort.Prepared | undefined;
}

interface PreparedItem<T> {
  value: T;
  prepared: Fuzzysort.Prepared | undefined;
}

export type PropertyGetter<T> = (item: T) => string;

/** A displayable search result item */
export interface SearchResult<T> {
  /** Undecorated string identifying the item */
  key: string;
  /** (Potentially) decorated string of HTML to be displayed */
  content: React.ReactNode;
  /** The raw value of the item */
  value: T;
}

export interface SearchableListStateArgs<T> {
  items: T[];
  propertyGetter: keyof T | PropertyGetter<T>;
}

function getProperty<T>(item: T, getter: keyof T | PropertyGetter<T>): string {
  return typeof getter === 'function' ? getter(item) : `${item[getter]}`;
}

/** Generates a plain list of un-highlighted search results */
function toSearchResults<T>(items: T[], getter: keyof T | PropertyGetter<T>): SearchResult<T>[] {
  return items.map((item) => {
    const property = getProperty(item, getter);
    return {
      content: property,
      key: property,
      value: item,
    };
  });
}

interface MatchRange {
  start: number;
  end: number;
}
const createHighlightedResult = <T extends {}>(
  result: Fuzzysort.KeyResult<SearchTarget<T>>,
  highlightElement = 'mark',
) => {
  const { indexes, target } = result;
  const parts: React.ReactNodeArray = [];
  let lastMatchRange: MatchRange = {
    start: 0,
    end: 0,
  };
  let prevIndex = 0;
  // fuzzysort generates an array of indices for each matching letter
  // in the target string. We want to wrap a highlight element around each
  // continuous range of matching indices. This will result in an array of
  // mixed fragments that are either string literals or an element with
  // a string literal as a child.
  indexes.forEach((matchIndex, idx) => {
    if (lastMatchRange.end !== matchIndex) {
      lastMatchRange = {
        start: matchIndex,
        end: matchIndex + 1,
      };
    } else {
      lastMatchRange.end = matchIndex + 1;
    }

    if (idx + 1 === indexes.length || indexes[idx + 1] !== lastMatchRange.end) {
      if (lastMatchRange.start !== 0) {
        parts.push(result.target.substring(prevIndex, lastMatchRange.start));
      }
      parts.push(
        createElement(
          highlightElement,
          { key: lastMatchRange.start },
          result.target.substring(lastMatchRange.start, lastMatchRange.end),
        ),
      );
      prevIndex = matchIndex + 1;
    }
  });

  if (lastMatchRange.end !== target.length) {
    parts.push(result.target.substring(lastMatchRange.end, target.length));
  }

  return createElement(Fragment, { children: parts });
};

/** Converts a prepared list of search targets into a list of search results */
function getFilteredItems<T>(targets: SearchTarget<T>[], searchString: string): SearchResult<T>[] {
  const results = fuzzysort.go(searchString, targets, {
    allowTypo: false,
    key: 'prepared',
  });
  return results.map((result) => {
    const content = createHighlightedResult(result);
    return {
      content,
      key: result.target,
      value: result.obj.value,
    };
  });
}

/** Manages state for fuzzy-matching a set of items by a search string, with the
 * resulting list of items being highlighted where characters in the search
 * string match characters in the item
 */
export const useSearchableListState = <T extends {}>({
  items,
  propertyGetter,
}: SearchableListStateArgs<T>) => {
  const [preparedItems, setPreparedItems] = useState<PreparedItem<T>[]>([]);
  const [results, setResults] = useState<SearchResult<T>[]>([]);
  const [searchString, setSearchString] = useState('');
  const [unfilteredResults, setUnfilteredResults] = useState<SearchResult<T>[]>([]);

  useEffect(() => {
    setUnfilteredResults(toSearchResults(items, propertyGetter));
    setPreparedItems(
      items.map((value) => ({
        value,
        prepared: fuzzysort.prepare(getProperty(value, propertyGetter)),
      })),
    );
  }, [items]);

  useEffect(() => {
    setResults(
      searchString.length === 0 ? unfilteredResults : getFilteredItems(preparedItems, searchString),
    );
  }, [preparedItems, unfilteredResults, searchString]);

  return {
    results,
    searchString,
    setSearchString,
  };
};
