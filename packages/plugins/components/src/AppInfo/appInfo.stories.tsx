import * as React from 'react';
import { ComponentStory, ComponentMeta } from '@storybook/react';
import { Card, CardContent } from '@material-ui/core';
import { AppInfo, INFO_WINDOW_WIDTH } from '.';
import { generateVersionInfo } from './appInfo.mock';
import { VersionDisplay } from './versionDisplay';

export default {
  title: 'Components/AppInfo',
  component: AppInfo,
} as ComponentMeta<typeof AppInfo>;

const Template: ComponentStory<typeof AppInfo> = (props) => <AppInfo {...props} />;

export const Default = Template.bind({});
Default.args = {
  versions: [
    generateVersionInfo('UI Version', '1.22.134'),
    generateVersionInfo('Admin Version', '3.2.1'),
    generateVersionInfo('Google Analytics', 'Active'),
  ],
  documentationUrl: 'here.is/some/link#',
};

export const ContentOnly = Template.bind({});
ContentOnly.args = {
  versions: [
    generateVersionInfo('Normal', '1.8.13'),
    generateVersionInfo('Long Version', 'Very Uncomfortable'),
    generateVersionInfo(
      'Long Name for all those who are interested in future endeavors',
      '1.23.12',
    ),
  ],
  documentationUrl: 'here.is/some/link#',
};
ContentOnly.decorators = [
  (_Story, context) => {
    return (
      <Card>
        <CardContent
          style={{
            display: 'flex',
            alignItems: 'center',
            flexDirection: 'column',
            width: `${INFO_WINDOW_WIDTH}px`,
          }}
        >
          <VersionDisplay {...context.args} />
        </CardContent>
      </Card>
    );
  },
];
