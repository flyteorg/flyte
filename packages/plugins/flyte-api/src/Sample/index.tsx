import * as React from 'react';
import { AppBar, Toolbar, IconButton, makeStyles, Theme } from '@material-ui/core';
import MenuIcon from '@material-ui/icons/Menu';

const useStyles = makeStyles((theme: Theme) => ({
  spacer: {
    flexGrow: 1,
  },
  menuButton: {
    marginRight: theme.spacing(2),
  },
}));

export interface SampleComponentProps {
  useCustomContent?: boolean; // rename to show that it is a backNavigation
  className?: string;
}

/** Contains all content in the top navbar of the application. */
export const SampleComponent = (props: SampleComponentProps) => {
  const styles = useStyles();

  return (
    <AppBar
      color="secondary"
      elevation={0}
      id="navbar"
      position="fixed"
      className={props.className as string}
    >
      <Toolbar>
        <div className={styles.spacer} />
        {' Sample Text '}
        <div className={styles.spacer} />
        <IconButton edge="start" className={styles.menuButton} color="inherit" aria-label="menu">
          <MenuIcon />
        </IconButton>
      </Toolbar>
    </AppBar>
  );
};
