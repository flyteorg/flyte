import { navBarContentId } from 'common/constants';
import { log } from 'common/log';
import * as React from 'react';
import * as ReactDOM from 'react-dom';

/** Complements NavBar, allowing pages to inject custom content. */
export const NavBarContent: React.FC<{}> = ({ children }) => {
  const navBar = document.getElementById(navBarContentId);
  if (navBar == null) {
    log.warn(`
            Attempting to mount content into NavBar, but failed to find the content component.
            Did you mount an instance of NavBar with useCustomContent=true?`);
    return null;
  }
  return ReactDOM.createPortal(children, navBar);
};
