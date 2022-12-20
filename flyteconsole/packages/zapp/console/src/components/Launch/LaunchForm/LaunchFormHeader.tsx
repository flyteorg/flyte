import { DialogTitle, Typography } from '@material-ui/core';
import * as React from 'react';
import { useStyles } from './styles';

interface LaunchFormHeaderProps {
  title?: string;
  formTitle: string;
}

/** Shared header component for the Launch form */
export const LaunchFormHeader: React.FC<LaunchFormHeaderProps> = ({ title = '', formTitle }) => {
  const styles = useStyles();
  return (
    <DialogTitle disableTypography={true} className={styles.header}>
      <div className={styles.inputLabel}>{formTitle}</div>
      <Typography variant="h6">{title}</Typography>
    </DialogTitle>
  );
};
