import { act, fireEvent, render, waitFor } from '@testing-library/react';
import * as React from 'react';
import { useDebouncedValue } from '../useDebouncedValue';

const inputLabel = 'input-box';
const debounceDelayMs = 10;

const DebounceTester = () => {
  const [value, setValue] = React.useState('');
  const debouncedValue = useDebouncedValue(value, debounceDelayMs);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    event.preventDefault();
    setValue(event.currentTarget.value);
  };

  return (
    <div>
      {debouncedValue}
      <input value={value} aria-label={inputLabel} onChange={handleChange} />
    </div>
  );
};

describe('useDebouncedValue', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.clearAllTimers();
    jest.useRealTimers();
  });
  const renderTester = () => render(<DebounceTester />);

  it('should delay state updates when receiving a new value', async () => {
    const { getByLabelText, queryByText } = renderTester();

    const inputBox = await waitFor(() => getByLabelText(inputLabel));
    const newValue = 'abcdefg';
    fireEvent.change(inputBox, { target: { value: newValue } });
    expect(queryByText(newValue)).toBeNull();
    act(() => {
      jest.advanceTimersByTime(debounceDelayMs);
    });
    expect(queryByText(newValue)).not.toBeNull();
  });

  it('should not generate a timer for initial value', () => {
    renderTester();
    // Clear any other timers that might have queued before a possible
    // debounce by running to debounceDelay - 1
    act(() => {
      jest.advanceTimersByTime(debounceDelayMs - 1);
    });
    expect(jest.getTimerCount()).toBe(0);
  });
});
