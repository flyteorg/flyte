import { render } from '@testing-library/react';
import * as React from 'react';

import { DataTable } from '../DataTable';

const mockData = { key1: 'value1', key2: 'value2' };

describe('DataTable', () => {
  it('should render a table with mocked data', () => {
    const { getAllByRole } = render(<DataTable data={mockData} />);
    const headers = getAllByRole('columnheader');
    const cells = getAllByRole('cell');
    expect(headers).toHaveLength(2);
    expect(headers[0]).toHaveTextContent('Key');
    expect(headers[1]).toHaveTextContent('Value');
    expect(cells).toHaveLength(4);
    expect(cells[0]).toHaveTextContent('key1');
    expect(cells[1]).toHaveTextContent('value1');
    expect(cells[2]).toHaveTextContent('key2');
    expect(cells[3]).toHaveTextContent('value2');
  });
});
