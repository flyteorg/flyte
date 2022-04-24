import { detailsPanelId } from 'common/constants';
import { log } from 'common/log';
import * as React from 'react';
import * as ReactDOM from 'react-dom';

/** Complements DetailsPanel, encapsulating the logic needed to locate
 * and create a portal into the DOM element. Children of this component will be
 * rendered into the DetailsPanel, if it exists.
 */
export const DetailsPanelContent: React.FC<{}> = ({ children }) => {
  const detailsPanel = document.getElementById(detailsPanelId);
  if (detailsPanel == null) {
    log.warn(`
            Attempting to mount content into DetailsPanel but it does not exist.
            Did you mount an instance of DetailsPanel?`);
    return null;
  }
  return ReactDOM.createPortal(children, detailsPanel);
};
