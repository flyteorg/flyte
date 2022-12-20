import { render, waitFor } from '@testing-library/react';
import * as React from 'react';

import { LoadingSpinner } from '../LoadingSpinner';

describe('LoadingSpinner', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.clearAllTimers();
    jest.useRealTimers();
  });

  it('delays rendering for 1 second', async () => {
    const { queryByTestId, getByTestId } = render(<LoadingSpinner />);
    await waitFor(() => {
      jest.advanceTimersByTime(500);
    });
    expect(queryByTestId('loading-spinner')).toBeNull();
    await waitFor(() => {
      jest.advanceTimersByTime(500);
    });
    expect(getByTestId('loading-spinner')).not.toBeNull();
  });
});
