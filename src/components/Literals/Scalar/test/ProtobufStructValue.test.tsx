import { render } from '@testing-library/react';
import * as React from 'react';

import { ProtobufStruct } from 'models';
import { ProtobufStructValue } from '../ProtobufStructValue';

describe('Scalars/ProtobufStructValue', () => {
    it('renders sorted keys', () => {
        const struct: ProtobufStruct = {
            fields: {
                input2: {
                    kind: 'nullValue'
                },
                input1: { kind: 'nullValue' }
            }
        };
        const { getAllByText } = render(
            <ProtobufStructValue struct={struct} />
        );
        const labels = getAllByText(/input/);
        expect(labels.length).toBe(2);
        expect(labels[0]).toHaveTextContent(/input1/);
        expect(labels[1]).toHaveTextContent(/input2/);
    });
});
