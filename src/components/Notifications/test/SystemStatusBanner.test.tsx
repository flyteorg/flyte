import { fireEvent, render, wait } from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext, APIContextValue } from 'components/data/apiContext';
import { StatusString, SystemStatus } from 'models';
import * as React from 'react';
import { SystemStatusBanner } from '../SystemStatusBanner';

describe('SystemStatusBanner', () => {
    let systemStatus: SystemStatus;
    let apiContext: APIContextValue;

    beforeEach(() => {
        systemStatus = {
            status: 'normal',
            message: 'Everything is fine.'
        };
        apiContext = mockAPIContextValue({
            getSystemStatus: jest
                .fn()
                .mockImplementation(() => Promise.resolve(systemStatus))
        });
    });

    const renderStatusBanner = () =>
        render(
            <APIContext.Provider value={apiContext}>
                <SystemStatusBanner />
            </APIContext.Provider>
        );

    it('should display an info icon for normal status', async () => {
        const { getByTestId } = renderStatusBanner();
        await wait();

        expect(getByTestId('info-icon')).toBeInTheDocument();
    });

    it('should display a waring icon for degraded status', async () => {
        systemStatus.status = 'degraded';
        const { getByTestId } = renderStatusBanner();
        await wait();

        expect(getByTestId('warning-icon')).toBeInTheDocument();
    });

    it('should display a warning icon for down status', async () => {
        systemStatus.status = 'down';
        const { getByTestId } = renderStatusBanner();
        await wait();

        expect(getByTestId('warning-icon')).toBeInTheDocument();
    });

    it('should render normal status icon for any unknown status', async () => {
        systemStatus.status = 'unknown' as StatusString;
        const { getByTestId } = renderStatusBanner();
        await wait();

        expect(getByTestId('info-icon')).toBeInTheDocument();
    });

    it('should only display if a `message` property is present', async () => {
        delete systemStatus.message;
        const { queryByRole } = renderStatusBanner();
        await wait();
        expect(queryByRole('banner')).toBeNull();
    });

    it('should hide when dismissed by user', async () => {
        const { getByRole, queryByRole } = renderStatusBanner();
        await wait();

        expect(getByRole('banner')).toBeInTheDocument();

        const closeButton = getByRole('button');
        fireEvent.click(closeButton);
        await wait();

        expect(queryByRole('banner')).toBeNull();
    });
});
