import * as React from 'react';
import { storiesOf } from '@storybook/react';
import Typography from '@material-ui/core/Typography';

const stories = storiesOf('Primitives/Material UI', module);
stories.add('Typography', () => (
  <div>
    <Typography variant="h5">H5 - Biggest header, not that we use it yet</Typography>
    <Typography variant="h6">H6 - Most used header</Typography>

    <Typography variant="body1">body 1 - for regular text</Typography>
    <Typography variant="body2">body 2 - looks like the same as body1?</Typography>

    <Typography variant="subtitle1">Subtitle1 - just some text for visibility</Typography>
    <Typography variant="overline">Overline - with some real text</Typography>
  </div>
));
