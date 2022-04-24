import { fireEvent, render } from '@testing-library/react';
import * as React from 'react';

import { SearchableList, SearchableListProps, SearchResult } from '../SearchableList';

const listStringItems: string[] = ['A Workflow', 'BetterWorkflow', 'ZZLastItemz'];

interface SimpleItem {
  id: string;
  label: string;
}
const listObjectItems: SimpleItem[] = listStringItems.map((s) => ({
  id: s,
  label: s,
}));

const renderContent = (results: SearchResult<SimpleItem>[]) => (
  <div>
    {results.map((r) => (
      <div role="list-item" aria-label={r.value.label} key={r.key}>
        {r.content}
      </div>
    ))}
  </div>
);

describe('SearchableList', () => {
  let props: SearchableListProps<SimpleItem>;
  const renderList = () => render(<SearchableList {...props} />);
  beforeEach(() => {
    props = {
      renderContent,
      items: [...listObjectItems],
      propertyGetter: 'id',
    };
  });

  it('should use custom placeholder if provided', () => {
    props.placeholder = 'custom placeholder';
    const { getByPlaceholderText } = renderList();
    expect(getByPlaceholderText(props.placeholder)).toBeTruthy();
  });

  describe('when a search string is provided', () => {
    // [input string, expected_matches[]]
    const searchCases: [string, string[]][] = [
      ['A', ['A Workflow', 'ZZLastItemz']],
      ['a', ['A Workflow', 'ZZLastItemz']], // should be case-insensitive
      ['B', ['BetterWorkflow']],
      ['C', []],
      ['W', ['A Workflow', 'BetterWorkflow']],
      ['Z', ['ZZLastItemz']],
      ['ZZ', ['ZZLastItemz']],
      ['ZZZ', ['ZZLastItemz']],
      ['ZZZZ', []],
    ];

    searchCases.forEach(([input, expectedValues]) => {
      const expectString = expectedValues.length
        ? `should match ${expectedValues} with input ${input}`
        : `should have no matches for input ${input}`;

      it(expectString, () => {
        const { getByRole, getByLabelText, queryAllByRole } = renderList();
        fireEvent.change(getByRole('search'), {
          target: { value: input },
        });

        expect(queryAllByRole('list-item').length).toEqual(expectedValues.length);
        expectedValues.forEach((value) => expect(getByLabelText(value)).toBeTruthy());
      });
    });

    it('should accept a string propertyGetter', () => {
      props.propertyGetter = 'id';
      const { getByRole, getByLabelText } = renderList();
      fireEvent.change(getByRole('search'), {
        target: { value: 'A W' },
      });

      expect(getByLabelText('A Workflow')).toBeTruthy();
    });

    it('should accept a function propertyGetter', () => {
      props.propertyGetter = (item) => item.id;
      const { getByRole, getByLabelText } = renderList();
      fireEvent.change(getByRole('search'), {
        target: { value: 'A W' },
      });

      expect(getByLabelText('A Workflow')).toBeTruthy();
    });
  });
});
