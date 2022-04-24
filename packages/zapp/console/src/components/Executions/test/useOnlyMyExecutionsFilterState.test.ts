import { renderHook } from '@testing-library/react-hooks';
import { loadedFetchable } from 'components/hooks/__mocks__/fetchableData';
import { useUserProfile } from 'components/hooks/useUserProfile';
import { useOnlyMyExecutionsFilterState } from 'components/Executions/filters/useOnlyMyExecutionsFilterState';
import { UserProfile } from 'models/Common/types';
import { FetchableData } from 'components/hooks/types';

jest.mock('components/hooks/useUserProfile');

describe('useOnlyMyExecutionsFilterState', () => {
  const mockUseUserProfile = useUserProfile as jest.Mock<FetchableData<UserProfile | null>>;
  mockUseUserProfile.mockReturnValue(loadedFetchable(null, jest.fn()));

  describe.each`
    isFilterDisabled | initialValue | expected
    ${undefined}     | ${undefined} | ${{ isFilterDisabled: false, onlyMyExecutionsValue: false }}
    ${false}         | ${undefined} | ${{ isFilterDisabled: false, onlyMyExecutionsValue: false }}
    ${true}          | ${undefined} | ${{ isFilterDisabled: true, onlyMyExecutionsValue: false }}
    ${undefined}     | ${false}     | ${{ isFilterDisabled: false, onlyMyExecutionsValue: false }}
    ${undefined}     | ${true}      | ${{ isFilterDisabled: false, onlyMyExecutionsValue: true }}
    ${false}         | ${false}     | ${{ isFilterDisabled: false, onlyMyExecutionsValue: false }}
    ${false}         | ${true}      | ${{ isFilterDisabled: false, onlyMyExecutionsValue: true }}
    ${true}          | ${false}     | ${{ isFilterDisabled: true, onlyMyExecutionsValue: false }}
    ${true}          | ${true}      | ${{ isFilterDisabled: true, onlyMyExecutionsValue: true }}
  `('for each case', ({ isFilterDisabled, initialValue, expected }) => {
    it(`should return ${JSON.stringify(
      expected,
    )} when called with isFilterDisabled = ${isFilterDisabled} and initialValue = ${initialValue}`, () => {
      const { result } = renderHook(() =>
        useOnlyMyExecutionsFilterState({ isFilterDisabled, initialValue }),
      );
      expect(result.current).toEqual(expect.objectContaining(expected));
    });
  });
});
