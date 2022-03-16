import { fireEvent, render, waitFor } from '@testing-library/react';
import { APIContext, APIContextValue } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { StatusString, SystemStatus } from 'models/Common/types';
import * as React from 'react';
import { pendingPromise } from 'test/utils';
import { SystemStatusBanner } from '../SystemStatusBanner';

describe('SystemStatusBanner', () => {
  let systemStatus: SystemStatus;
  let apiContext: APIContextValue;
  let getSystemStatus: jest.Mock<ReturnType<APIContextValue['getSystemStatus']>>;

  beforeEach(() => {
    systemStatus = {
      status: 'normal',
      message: 'Everything is fine.',
    };
    getSystemStatus = jest.fn().mockImplementation(() => Promise.resolve(systemStatus));
    apiContext = mockAPIContextValue({
      getSystemStatus,
    });
  });

  const renderStatusBanner = () =>
    render(
      <APIContext.Provider value={apiContext}>
        <SystemStatusBanner />
      </APIContext.Provider>,
    );

  it('should display an info icon for normal status', async () => {
    const { getByTestId } = renderStatusBanner();
    await waitFor(() => {});

    expect(getByTestId('info-icon')).toBeInTheDocument();
  });

  it('should display a warning icon for degraded status', async () => {
    systemStatus.status = 'degraded';
    const { getByTestId } = renderStatusBanner();
    await waitFor(() => {});

    expect(getByTestId('warning-icon')).toBeInTheDocument();
  });

  it('should display a warning icon for down status', async () => {
    systemStatus.status = 'down';
    const { getByTestId } = renderStatusBanner();
    await waitFor(() => {});

    expect(getByTestId('warning-icon')).toBeInTheDocument();
  });

  it('should render normal status icon for any unknown status', async () => {
    systemStatus.status = 'unknown' as StatusString;
    const { getByTestId } = renderStatusBanner();
    await waitFor(() => {});

    expect(getByTestId('info-icon')).toBeInTheDocument();
  });

  it('should only display if a `message` property is present', async () => {
    delete systemStatus.message;
    const { queryByRole } = renderStatusBanner();
    await waitFor(() => {});
    expect(queryByRole('banner')).toBeNull();
  });

  it('should hide when dismissed by user', async () => {
    const { getByRole, queryByRole } = renderStatusBanner();
    await waitFor(() => {});

    expect(getByRole('banner')).toBeInTheDocument();

    const closeButton = getByRole('button');
    fireEvent.click(closeButton);
    await waitFor(() => {});

    expect(queryByRole('banner')).toBeNull();
  });

  it('should render inline urls as links', async () => {
    systemStatus.message = 'Check out http://flyte.org for more info';
    const { getByText } = renderStatusBanner();
    await waitFor(() => {});
    expect(getByText('http://flyte.org').closest('a')).toHaveAttribute('href', 'http://flyte.org');
  });

  it('should render no content while loading', async () => {
    getSystemStatus.mockImplementation(() => pendingPromise());
    const { container } = renderStatusBanner();
    await waitFor(() => {});
    expect(container.hasChildNodes()).toBe(false);
  });

  it('should render no content on error', async () => {
    getSystemStatus.mockRejectedValue('Failed');
    const { container } = renderStatusBanner();
    await waitFor(() => {});
    expect(container.hasChildNodes()).toBe(false);
  });
});
