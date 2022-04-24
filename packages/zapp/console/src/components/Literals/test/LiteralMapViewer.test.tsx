import { render } from '@testing-library/react';
import { LiteralMap } from 'models/Common/types';
import * as React from 'react';
import { LiteralMapViewer } from '../LiteralMapViewer';

describe('LiteralMapViewer', () => {
  it('renders sorted keys', () => {
    const literals: LiteralMap = {
      literals: {
        input2: {},
        input1: {},
      },
    };
    const { getAllByText } = render(<LiteralMapViewer map={literals} />);
    const labels = getAllByText(/input/);
    expect(labels.length).toBe(2);
    expect(labels[0]).toHaveTextContent(/input1/);
    expect(labels[1]).toHaveTextContent(/input2/);
  });
});
