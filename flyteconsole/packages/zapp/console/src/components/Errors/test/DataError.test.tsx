import { render } from '@testing-library/react';
import { NotAuthorizedError, NotFoundError } from 'errors/fetchErrors';
import * as React from 'react';
import { DataError, DataErrorProps } from '../DataError';

describe('DataError', () => {
  const defaultProps: DataErrorProps = {
    errorTitle: 'Test Error',
  };

  it('renders nothing for NotAuthorized errors', () => {
    const { container } = render(<DataError {...defaultProps} error={new NotAuthorizedError()} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders not found for NotFound errors', () => {
    const { getByText } = render(<DataError {...defaultProps} error={new NotFoundError('')} />);
    expect(getByText('Not found')).not.toBeEmptyDOMElement();
  });
});
