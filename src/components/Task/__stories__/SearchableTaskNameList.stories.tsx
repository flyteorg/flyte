import { Collapse } from '@material-ui/core';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { sampleTaskNames } from 'models/__mocks__/sampleTaskNames';
import { SnackbarProvider } from 'notistack';
import * as React from 'react';
import { SearchableTaskNameList } from '../SearchableTaskNameList';

const baseProps = { names: [...sampleTaskNames] };

// wrapper - to ensure that error/success notification shown as expected in storybook
const Wrapper = (props) => {
  return (
    <SnackbarProvider
      maxSnack={2}
      anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
      TransitionComponent={Collapse}
    >
      {props.children}
    </SnackbarProvider>
  );
};

const stories = storiesOf('Task/SearchableTaskNameList', module);
stories.addDecorator((story) => <div style={{ width: '650px' }}>{story()}</div>);
stories.add('basic', () => (
  <Wrapper>
    <SearchableTaskNameList
      onArchiveFilterChange={action('onArchiveFilterChange')}
      showArchived={false}
      {...baseProps}
    />
  </Wrapper>
));
