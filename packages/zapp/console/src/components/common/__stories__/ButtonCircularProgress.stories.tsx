import { Button } from '@material-ui/core';
import { storiesOf } from '@storybook/react';
import * as React from 'react';

import { ButtonCircularProgress } from '../ButtonCircularProgress';

const ButtonWithProgress: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  React.useEffect(() => {
    if (loading) {
      const timeout = setTimeout(() => setLoading(false), 2000);
      return () => clearTimeout(timeout);
    }
    return;
  }, [loading]);
  const onClick = () => setLoading(true);
  return (
    <div>
      <Button variant="contained" color="primary" disabled={loading} onClick={onClick}>
        Click to Load
        {loading && <ButtonCircularProgress />}
      </Button>
    </div>
  );
};

const stories = storiesOf('Common', module);
stories.add('ButtonCircularProgress', () => <ButtonWithProgress />);
