import { Collapse } from '@material-ui/core';
import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { sampleWorkflowNames } from 'models/__mocks__/sampleWorkflowNames';
import { SnackbarProvider } from 'notistack';
import * as React from 'react';
import { SearchableWorkflowNameList } from '../SearchableWorkflowNameList';

const baseProps = { workflows: [...sampleWorkflowNames] };

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

const stories = storiesOf('Workflow/SearchableWorkflowNameList', module);
stories.addDecorator((story) => <div style={{ width: '650px' }}>{story()}</div>);
stories.add('basic', () => (
  <Wrapper>
    <SearchableWorkflowNameList
      showArchived={false}
      onArchiveFilterChange={action('onArchiveFilterChange')}
      {...baseProps}
    />
  </Wrapper>
));
