import IconButton from '@material-ui/core/IconButton';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import MoreVert from '@material-ui/icons/MoreVert';
import * as React from 'react';
import { labels } from './constants';

export interface MoreOptionsMenuItem {
  label: string;
  onClick: () => void;
}

export interface MoreOptionsMenuProps {
  className?: string;
  options: MoreOptionsMenuItem[];
}
/** Renders a vertical three-dots menu button with the provided options.
 * Each option should have a label and corresponding onClick handler, which will
 * be invoked when the item is clicked.
 */
export const MoreOptionsMenu: React.FC<MoreOptionsMenuProps> = ({ className, options }) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const { handleClose, handleClickMenuButton, listItems } = React.useMemo(() => {
    const handleClickMenuButton = (event: React.MouseEvent<HTMLButtonElement>) => {
      setAnchorEl(event.currentTarget);
    };

    const handleClose = () => {
      setAnchorEl(null);
    };

    const listItems = options.map(({ label, onClick: handleItemClick }) => {
      const onClick = () => {
        setAnchorEl(null);
        handleItemClick();
      };

      return (
        <MenuItem key={label} onClick={onClick}>
          {label}
        </MenuItem>
      );
    });

    return { handleClickMenuButton, handleClose, listItems };
  }, [options, setAnchorEl]);

  return (
    <div className={className}>
      <IconButton
        aria-controls="more-options-menu"
        aria-haspopup="true"
        aria-label={labels.moreOptionsButton}
        color="inherit"
        onClick={handleClickMenuButton}
      >
        <MoreVert />
      </IconButton>
      <Menu
        aria-label={labels.moreOptionsMenu}
        id="more-options-menu"
        anchorEl={anchorEl}
        open={!!anchorEl}
        onClose={handleClose}
      >
        {listItems}
      </Menu>
    </div>
  );
};
