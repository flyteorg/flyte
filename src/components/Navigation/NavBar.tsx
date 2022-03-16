import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import { navBarContentId } from 'common/constants';
import * as React from 'react';
import { DefaultAppBarContent } from './DefaultAppBarContent';

/** Contains all content in the top navbar of the application. */
export const NavBar: React.FC<{ useCustomContent?: boolean }> = ({
  /** Allow injection of custom content. */
  useCustomContent = false,
}) => {
  const content = useCustomContent ? <div id={navBarContentId} /> : <DefaultAppBarContent />;
  return (
    <AppBar color="secondary" elevation={0} id="navbar" position="fixed">
      <Toolbar id={navBarContentId}>{content}</Toolbar>
    </AppBar>
  );
};
