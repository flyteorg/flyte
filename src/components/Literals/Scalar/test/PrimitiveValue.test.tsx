import { render } from '@testing-library/react';
import { Primitive } from 'models/Common/types';
import * as React from 'react';
import { long } from 'test/utils';
import { PrimitiveValue } from '../PrimitiveValue';

describe('PrimitiveValue', () => {
  it('renders datetime', () => {
    const primitive: Primitive = {
      value: 'datetime',
      datetime: {
        seconds: long(3600),
        nanos: 0,
      },
      boolean: false,
      integer: long(0),
      floatValue: 0,
      stringValue: '',
    };
    const { getByText } = render(<PrimitiveValue primitive={primitive} />);
    expect(getByText('1/1/1970 1:00:00 AM UTC')).toBeInTheDocument();
  });
});
