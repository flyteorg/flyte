import { DialogTitle, IconButton, Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Close from '@material-ui/icons/Close';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    margin: 0,
    padding: theme.spacing(2),
  },
  closeButton: {
    color: theme.palette.text.primary,
    position: 'absolute',
    right: theme.spacing(1),
    top: theme.spacing(1),
  },
}));

export interface ClosableDialogTitleProps {
  children: React.ReactNode;
  onClose: () => void;
}

/** A replacement for MUI's DialogTitle which also renders a close button */
export const ClosableDialogTitle: React.FC<ClosableDialogTitleProps> = ({ children, onClose }) => {
  const styles = useStyles();
  return (
    <DialogTitle disableTypography={true} className={styles.root}>
      <Typography variant="h6">{children}</Typography>
      {onClose ? (
        <IconButton aria-label="Close" className={styles.closeButton} onClick={onClose}>
          <Close />
        </IconButton>
      ) : null}
    </DialogTitle>
  );
};
