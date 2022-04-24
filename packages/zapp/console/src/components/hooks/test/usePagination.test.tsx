import { fireEvent, getByLabelText, getByText, render, waitFor } from '@testing-library/react';
import { PaginatedEntityResponse, RequestConfig } from 'models/AdminEntity/types';
import * as React from 'react';
import { PaginationConfig, usePagination } from '../usePagination';

const valueLabel = 'pagination-value';
const fetchLabel = 'pagination-doFetch';
const moreItemsAvailableLabel = 'pagination-moreItemsAvailable';

interface PaginationItem {
  id: string;
}

type FetchResponse = PaginatedEntityResponse<PaginationItem>;
interface PaginationTesterProps {
  config: PaginationConfig<{}>;
  doFetch: jest.Mock<Promise<FetchResponse>>;
}

const PaginationTester = ({ config, doFetch }: PaginationTesterProps) => {
  const fetchable = usePagination(config, doFetch);
  const onClickFetch = () => fetchable.fetch();

  return (
    <div>
      <div aria-label={valueLabel}>
        <ul>
          {fetchable.value.map(({ id }) => (
            <li key={`item-${id}`}>{`item-${id}`}</li>
          ))}
        </ul>
      </div>
      <div aria-label={moreItemsAvailableLabel}>
        {fetchable.moreItemsAvailable ? 'true' : 'false'}
      </div>
      <button aria-label={fetchLabel} onClick={onClickFetch}>
        Fetch Data
      </button>
    </div>
  );
};

describe('usePagination', () => {
  let entityCounter: number;
  let config: PaginationConfig<{}>;
  let doFetch: jest.Mock<Promise<FetchResponse>>;

  beforeEach(() => {
    entityCounter = 0;
    doFetch = jest.fn().mockImplementation((fetchArg: any, { limit = 25 }: RequestConfig) =>
      Promise.resolve({
        entities: Array.from({ length: limit }, () => {
          const id = `${entityCounter}`;
          entityCounter += 1;
          return { id };
        }),
        token: `${entityCounter}`,
      }),
    );
    config = {
      cacheItems: false,
      fetchArg: {},
      limit: 25,
    };
  });

  const renderTester = () => render(<PaginationTester config={config} doFetch={doFetch} />);
  const getElements = async (container: HTMLElement) => {
    return waitFor(() => {
      return {
        fetchButton: getByLabelText(container, fetchLabel),
        moreItemsAvailable: getByLabelText(container, moreItemsAvailableLabel),
      };
    });
  };

  const waitForLastItemRendered = async (container: HTMLElement) => {
    return waitFor(() => getByText(container, `item-${entityCounter - 1}`));
  };

  it('should pass returned token in subsequent calls', async () => {
    const { container } = renderTester();
    const { fetchButton } = await getElements(container);

    expect(doFetch).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({ token: '' }));

    fireEvent.click(fetchButton);
    await waitForLastItemRendered(container);
    expect(doFetch).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ token: `${config.limit}` }),
    );
  });

  it('should reset token when config changes', async () => {
    const { container } = renderTester();
    await getElements(container);

    expect(doFetch).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({ token: '' }));

    doFetch.mockClear();
    entityCounter = 0;

    // Change the config to trigger a rest of the pagination hook
    config.limit = 10;
    await getElements(renderTester().container);

    expect(doFetch).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({ token: '' }));
  });

  it('should set moreItemsAvailable if token is returned', async () => {
    const { moreItemsAvailable } = await getElements(renderTester().container);
    expect(moreItemsAvailable.textContent).toBe('true');
  });

  it('should not set moreItemsAvailable if no token is returned', async () => {
    doFetch.mockResolvedValue({ entities: [{ id: '0' }] });
    const { moreItemsAvailable } = await getElements(renderTester().container);
    expect(moreItemsAvailable.textContent).toBe('false');
  });
});
