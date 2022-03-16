import { render } from '@testing-library/react';
import { FetchableData } from 'components/hooks/types';
import { useRemoteLiteralMap } from 'components/hooks/useRemoteLiteralMap';
import { loadedFetchable } from 'components/hooks/__mocks__/fetchableData';
import * as Long from 'long';
import { LiteralMap } from 'models/Common/types';
import * as React from 'react';
import { RemoteLiteralMapViewer } from '../RemoteLiteralMapViewer';

jest.mock('components/hooks/useRemoteLiteralMap');

describe('RemoteLiteralMapViewer', () => {
  it('renders no data available', () => {
    const blob = {
      url: '',
      bytes: Long.fromInt(0),
    };

    const { getAllByText } = render(<RemoteLiteralMapViewer map={null} blob={blob} />);

    const items = getAllByText('No data is available.');
    expect(items.length).toBe(1);
  });

  it('renders map if it is defined', () => {
    const blob = {
      url: 'http://url',
      bytes: Long.fromInt(1337),
    };

    const map: LiteralMap = {
      literals: {
        input1: {},
      },
    };

    const { getAllByText } = render(<RemoteLiteralMapViewer map={map} blob={blob} />);

    const items = getAllByText('input1:');
    expect(items.length).toBe(1);
  });

  it('fetches blob if map is null', () => {
    const map: LiteralMap = {
      literals: {
        input1: {},
      },
    };

    const mockUseRemoteLiteralMap = useRemoteLiteralMap as jest.Mock<FetchableData<LiteralMap>>;
    mockUseRemoteLiteralMap.mockReturnValue(loadedFetchable(map, () => null));

    const blob = {
      url: 'http://url',
      bytes: Long.fromInt(1337),
    };

    const { getAllByText } = render(<RemoteLiteralMapViewer map={null} blob={blob} />);

    const items = getAllByText('input1:');
    expect(items.length).toBe(1);
  });
});
