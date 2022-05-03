import * as React from 'react';

import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
// import { navBarContentId } from 'common/constants';
import { DefaultNavBarContent } from './defaultContent';

export const navBarContentId = 'nav-bar-content';

interface NavBarProps {
  useCustomContent?: boolean; // rename to show that it is a backNavigation
  className?: string;
}

/** Contains all content in the top navbar of the application. */
export const NavBar = (props: NavBarProps) => {
  // const content = props.useCustomContent ? <div id={navBarContentId} /> : <DefaultNavBarContent />;
  return (
    <AppBar
      color="secondary"
      elevation={0}
      id="navbar"
      position="fixed"
      className={props.className as string}
    >
      <Toolbar id={navBarContentId}>
        {/* {content} */}
        <DefaultNavBarContent />
      </Toolbar>
    </AppBar>
  );
};

// export const NavBar = (): React.ReactElement => {
//   return <div>It&apos;s me - Navigation Bar</div>;
// };

export const add = (a: number, b: number) => {
  return a + b;
};
