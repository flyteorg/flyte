import { fireEvent, render, wait } from '@testing-library/react';
import * as React from 'react';

import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import {
    APIContext,
    APIContextValue,
    useAPIState
} from 'components/data/apiContext';
import { useFetchableData } from 'components/hooks';
import { NotAuthorizedError } from 'errors';
import { getUserProfile, UserProfile } from 'models';
import { UserInformation } from '../UserInformation';

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
            <UserInformation />
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

describe('UserInformation', () => {
    const sampleUserProfile: UserProfile = {
        preferredUsername: 'testUser@example.com'
    } as UserProfile;

    let mockGetUserProfile: jest.Mock<ReturnType<typeof getUserProfile>>;

    const UserInformationWithContext = () => (
        <APIContext.Provider
            value={mockAPIContextValue({ getUserProfile: mockGetUserProfile })}
        >
            <UserInformation />
        </APIContext.Provider>
    );

    beforeEach(() => {
        mockGetUserProfile = jest.fn().mockResolvedValue(null);
    });

    it('Shows login link if no user profile exists', async () => {
        const { getByText } = render(<UserInformationWithContext />);
        await wait(() => getByText('Login'));
        expect(mockGetUserProfile).toHaveBeenCalled();
        const element = getByText('Login');
        expect(element).toBeInTheDocument();
        expect(element.tagName).toBe('A');
    });

    it('Shows user preferredName if profile exists', async () => {
        mockGetUserProfile.mockResolvedValue(sampleUserProfile);
        const { getByText } = render(<UserInformationWithContext />);
        await wait(() => getByText(sampleUserProfile.preferredUsername));
        expect(mockGetUserProfile).toHaveBeenCalled();
        expect(
            getByText(sampleUserProfile.preferredUsername)
        ).toBeInTheDocument();
    });
});
