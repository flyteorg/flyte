import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { MenuItem, Select } from '@material-ui/core';
import { useHistory } from 'react-router-dom';
import { makeRoute } from 'routes/utils';
import { headerFontFamily } from 'components/Theme/constants';
import { FlyteNavItem } from './utils';

const useStyles = makeStyles((theme: Theme) => ({
  selectStyling: {
    minWidth: '120px',
    margin: theme.spacing(0, 2),
    color: 'inherit',
    '&:hover': {
      color: 'inherit',
    },
    '&:before, &:after, &:not(.Mui-disabled):hover::before': {
      border: 'none',
    },
  },
  colorInherit: {
    color: 'inherit',
    fontFamily: headerFontFamily,
    fontWeight: 600,
    lineHeight: 1.75,
  },
}));

interface NavigationDropdownProps {
  items: FlyteNavItem[]; // all other navigation items
  console?: string; // name for default navigation, if not provided "Console" is used.
}

/** Renders the default content for the app bar, which is the logo and help links */
export const NavigationDropdown = (props: NavigationDropdownProps) => {
  // Flyte Console list item - always there ans is first in the list
  const ConsoleItem: FlyteNavItem = React.useMemo(() => {
    return {
      title: props.console ?? 'Console',
      url: makeRoute('/'),
    };
  }, [props.console]);

  const [selectedPage, setSelectedPage] = React.useState<string>(ConsoleItem.title);
  const [open, setOpen] = React.useState(false);

  const history = useHistory();
  const styles = useStyles();

  const handleItemSelection = (item: FlyteNavItem) => {
    setSelectedPage(item.title);

    if (item.url.startsWith('+')) {
      // local navigation with BASE_URL addition
      history.push(makeRoute(item.url.slice(1)));
    } else {
      // treated as external navigation
      window.location.assign(item.url);
    }
  };

  return (
    <Select
      labelId="demo-controlled-open-select-label"
      id="demo-controlled-open-select"
      open={open}
      onClose={() => setOpen(false)}
      onOpen={() => setOpen(true)}
      value={selectedPage}
      className={styles.selectStyling}
      classes={{
        // update color of text and icon
        root: styles.colorInherit,
        icon: styles.colorInherit,
      }}
      MenuProps={{
        anchorOrigin: {
          vertical: 'bottom',
          horizontal: 'left',
        },
        transformOrigin: {
          vertical: 'top',
          horizontal: 'left',
        },
        getContentAnchorEl: null,
      }}
    >
      <MenuItem
        key={ConsoleItem.title}
        value={ConsoleItem.title}
        onClick={() => handleItemSelection(ConsoleItem)}
      >
        {ConsoleItem.title}
      </MenuItem>
      {props.items.map((item) => (
        <MenuItem key={item.title} value={item.title} onClick={() => handleItemSelection(item)}>
          {item.title}
        </MenuItem>
      ))}
    </Select>
  );
};
