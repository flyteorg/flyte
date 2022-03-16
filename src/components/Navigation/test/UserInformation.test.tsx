import { render, waitFor } from '@testing-library/react';
import { APIContext } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { getUserProfile } from 'models/Common/api';
import { UserProfile } from 'models/Common/types';
import * as React from 'react';
import { UserInformation } from '../UserInformation';

describe('UserInformation', () => {
  const sampleUserProfile: UserProfile = {
    preferredUsername: 'testUser@example.com',
  } as UserProfile;

  let mockGetUserProfile: jest.Mock<ReturnType<typeof getUserProfile>>;

  const UserInformationWithContext = () => (
    <APIContext.Provider value={mockAPIContextValue({ getUserProfile: mockGetUserProfile })}>
      <UserInformation />
    </APIContext.Provider>
  );

  beforeEach(() => {
    mockGetUserProfile = jest.fn().mockResolvedValue(null);
  });

  it('Shows login link if no user profile exists', async () => {
    const { getByText } = render(<UserInformationWithContext />);
    await waitFor(() => getByText('Login'));
    expect(mockGetUserProfile).toHaveBeenCalled();
    const element = getByText('Login');
    expect(element).toBeInTheDocument();
    expect(element.tagName).toBe('A');
  });

  it('Shows user preferredName if profile exists', async () => {
    mockGetUserProfile.mockResolvedValue(sampleUserProfile);
    const { getByText } = render(<UserInformationWithContext />);
    await waitFor(() => getByText(sampleUserProfile.preferredUsername));
    expect(mockGetUserProfile).toHaveBeenCalled();
    expect(getByText(sampleUserProfile.preferredUsername)).toBeInTheDocument();
  });
});
