import { fireEvent, render, waitFor } from '@testing-library/react';
import * as React from 'react';

import { APIContext, useAPIState } from 'components/data/apiContext';
import { useFetchableData } from 'components/hooks';
import { NotAuthorizedError } from 'errors';
import { LoginExpiredHandler } from '../LoginExpiredHandler';

function useTriggerNotAuthorizedFetchable() {
    return useFetchableData<{}>({
        defaultValue: {},
        autoFetch: false,
        doFetch: () => Promise.reject(new NotAuthorizedError())
    });
}

const LoginExpiredContent: React.FC = () => {
    const fetchable = useTriggerNotAuthorizedFetchable();
    const onClick = () => fetchable.fetch();
    return (
        <>
            <button onClick={onClick}>Trigger</button>
            <LoginExpiredHandler />
        </>
    );
};

const LoginExpiredContainer: React.FC = () => {
    const apiState = useAPIState();
    return (
        <APIContext.Provider value={apiState}>
            <LoginExpiredContent />
        </APIContext.Provider>
    );
};

describe('LoginExpiredHandler', () => {
    it('appears after an API request returns 401', async () => {
        const { getByText, queryByText } = render(<LoginExpiredContainer />);
        expect(queryByText('Login')).toBeNull();

        fireEvent(
            getByText('Trigger'),
            new MouseEvent('click', {
                bubbles: true,
                cancelable: true
            })
        );

        await waitFor(() => getByText('Login'));
    });
});
