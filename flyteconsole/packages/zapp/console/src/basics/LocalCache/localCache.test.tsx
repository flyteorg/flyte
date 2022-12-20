import * as React from 'react';
import { render, screen, act, fireEvent, getByTestId } from '@testing-library/react';
import { ClearLocalCache, LocalCacheItem, useLocalCache, onlyForTesting } from '.';
import { LocalCacheProvider } from './ContextProvider';

const SHOW_TEXT = 'SHOWED';
const HIDDEN_TEXT = 'HIDDEN';

const TestFrame = () => {
  const [show, setShow, clearShow] = useLocalCache(LocalCacheItem.TestSettingBool);

  const toShow = () => setShow(true);
  const toClear = () => clearShow();
  return (
    <div>
      <div>{show ? SHOW_TEXT : HIDDEN_TEXT}</div>
      <button onClick={toShow} data-testid="show" />
      <button onClick={toClear} data-testid="clear" />
    </div>
  );
};

const TestPage = () => {
  return (
    <LocalCacheProvider>
      <TestFrame />
    </LocalCacheProvider>
  );
};

describe('LocalCache', () => {
  beforeAll(() => {
    ClearLocalCache();
  });

  afterAll(() => {
    ClearLocalCache();
  });

  it('Can be used by component as expected', () => {
    const { container } = render(<TestPage />);

    const show = getByTestId(container, 'show');
    const clear = getByTestId(container, 'clear');

    expect(screen.getByText(HIDDEN_TEXT)).toBeTruthy();

    // change value
    act(() => {
      fireEvent.click(show);
    });
    expect(screen.getByText(SHOW_TEXT)).toBeTruthy();

    // reset to default
    act(() => {
      fireEvent.click(clear);
    });
    expect(screen.getByText(HIDDEN_TEXT)).toBeTruthy();
  });

  it('With no default value - assumes null', () => {
    const { getDefault } = onlyForTesting;
    expect(getDefault(LocalCacheItem.TestUndefined)).toBeNull();
  });

  it('Can store use default as an object', () => {
    const { getDefault } = onlyForTesting;
    expect(getDefault(LocalCacheItem.TestObject)).toMatchObject({
      name: 'Stella',
      age: '125',
    });
  });
});
