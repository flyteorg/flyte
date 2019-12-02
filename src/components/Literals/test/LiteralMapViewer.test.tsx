import { render } from '@testing-library/react';
import * as React from 'react';

import { LiteralMap } from 'models';
import { LiteralMapViewer } from '../LiteralMapViewer';

describe('Literals/LiteralMapViewer', () => {
    it('renders sorted keys', () => {
        const literals: LiteralMap = {
            literals: {
                input2: {},
                input1: {}
            }
        };
        const { getAllByText } = render(<LiteralMapViewer map={literals} />);
        const labels = getAllByText(/input/);
        expect(labels.length).toBe(2);
        expect(labels[0]).toHaveTextContent(/input1/);
        expect(labels[1]).toHaveTextContent(/input2/);
    });
});
