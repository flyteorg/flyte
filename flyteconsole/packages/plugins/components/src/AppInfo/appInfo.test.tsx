import * as React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react';

import { VersionDisplay } from './versionDisplay';
import { AppInfo } from '.';
import t from './strings';
import { generateVersionInfo } from './appInfo.mock';

describe('appInfo', () => {
  it('Info icon opens VersionDisplay on click', async () => {
    const title = t('modalTitle');
    const { getByTestId, queryByText } = render(<AppInfo versions={[]} documentationUrl="#" />);

    // click on the icon to open modal
    const infoIcon = getByTestId('infoIcon');
    fireEvent.click(infoIcon);
    await waitFor(() => {
      expect(queryByText(title)).toBeInTheDocument();
    });

    // click on close button should close modal
    const closeButton = getByTestId('closeButton');
    fireEvent.click(closeButton);
    await waitFor(() => expect(queryByText(title)).not.toBeInTheDocument());
  });

  it('VersionDisplay shows provided versions', async () => {
    const versions = [
      generateVersionInfo('UI Version', 'Active'),
      generateVersionInfo('Admin Version', '3.2.112'),
    ];

    const { queryByText } = render(<VersionDisplay versions={versions} documentationUrl="#" />);

    // click on the icon to open modal
    await waitFor(() => {
      expect(queryByText('UI Version')).toBeInTheDocument();
      expect(queryByText('Active')).toBeInTheDocument();

      expect(queryByText('Admin Version')).toBeInTheDocument();
      expect(queryByText('3.2.112')).toBeInTheDocument();
    });
  });
});
