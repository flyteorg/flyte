import { makeStyles, Theme, Typography } from '@material-ui/core';
import * as React from 'react';
import t from './strings';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    color: theme.palette.text.disabled,
    display: 'flex',
    justifyContent: 'center',
    marginTop: theme.spacing(4),
  },
}));

export const NoResults: React.FC = () => (
  <Typography className={useStyles().container} variant="h6" component="div">
    {t('noMatchingResults')}
  </Typography>
);
