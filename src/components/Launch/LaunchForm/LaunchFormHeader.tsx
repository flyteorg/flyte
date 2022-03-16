import { DialogTitle, Typography } from '@material-ui/core';
import * as React from 'react';
import { formStrings } from './constants';
import { useStyles } from './styles';

/** Shared header component for the Launch form */
export const LaunchFormHeader: React.FC<{ title?: string }> = ({ title = '' }) => {
  const styles = useStyles();
  return (
    <DialogTitle disableTypography={true} className={styles.header}>
      <div className={styles.inputLabel}>{formStrings.title}</div>
      <Typography variant="h6">{title}</Typography>
    </DialogTitle>
  );
};
