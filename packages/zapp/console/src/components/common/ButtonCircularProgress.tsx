import { CircularProgress } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import * as React from 'react';

const useStyles = makeStyles({
  root: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    marginTop: -12,
    marginLeft: -12,
  },
});

export const ButtonCircularProgress: React.FC<{}> = () => {
  const styles = useStyles();
  return <CircularProgress size={24} className={styles.root} />;
};
