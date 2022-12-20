import { Drawer, Paper } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { CSSProperties } from '@material-ui/styles';
import { detailsPanelId } from 'common/constants';
import { useTheme } from 'components/Theme/useTheme';
import * as React from 'react';
import { detailsPanelWidth } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
  modal: {
    pointerEvents: 'none',
  },
  paper: {
    display: 'flex',
    flex: '1 1 100%',
    maxHeight: '100%',
    paddingBottom: theme.spacing(2),
    pointerEvents: 'initial',
    width: detailsPanelWidth,
  },
  spacer: theme.mixins.toolbar as CSSProperties,
}));

interface DetailsPanelProps {
  onClose?: () => void;
  open?: boolean;
}

/** A shared panel rendered along the right side of the UI. Content can be
 * rendered into it using `DetailsPanelContent`
 */
export const DetailsPanel: React.FC<DetailsPanelProps> = ({ children, onClose, open = false }) => {
  const styles = useStyles();
  const theme = useTheme();
  return (
    <Drawer
      anchor="right"
      data-testid="details-panel"
      ModalProps={{
        className: styles.modal,
        hideBackdrop: true,
        // Modal uses inline styling for the zIndex, so we have to
        // override it
        style: { zIndex: theme.zIndex.appBar - 2 },
      }}
      onClose={onClose}
      open={open}
    >
      <div className={styles.spacer} />
      <Paper className={styles.paper} id={detailsPanelId} square={true}>
        {children}
      </Paper>
    </Drawer>
  );
};
