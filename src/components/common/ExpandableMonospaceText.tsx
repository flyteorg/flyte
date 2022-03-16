import { Button, ButtonBase } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import FileCopy from '@material-ui/icons/FileCopy';
import classnames from 'classnames';
import {
  bodyFontFamily,
  errorBackgroundColor,
  listhoverColor,
  nestedListColor,
  separatorColor,
  smallFontSize,
} from 'components/Theme/constants';
import * as copyToClipboard from 'copy-to-clipboard';
import * as React from 'react';

const expandedClass = 'expanded';
const showOnHoverClass = 'showOnHover';

export const useExpandableMonospaceTextStyles = makeStyles((theme: Theme) => ({
  actionContainer: {
    transition: theme.transitions.create('opacity', {
      duration: theme.transitions.duration.shorter,
      easing: theme.transitions.easing.easeInOut,
    }),
  },
  bottomFade: {
    background: 'linear-gradient(rgba(255,252,252,0.39), rgba(255,255,255,0.79))',
    bottom: 0,
    height: theme.spacing(4),
    left: 0,
    position: 'absolute',
    right: 0,
  },
  container: {
    backgroundColor: errorBackgroundColor,
    border: `1px solid ${separatorColor}`,
    borderRadius: 4,
    height: theme.spacing(12),
    minHeight: theme.spacing(12),
    overflowY: 'hidden',
    padding: theme.spacing(2),
    position: 'relative',
    '$nestedParent &': {
      backgroundColor: theme.palette.common.white,
    },
    [`&.${expandedClass}`]: {
      height: 'auto',
    },
    // All children using the showOnHover class will be hidden until
    // the mouse enters the container
    [`& .${showOnHoverClass}`]: {
      opacity: 0,
    },
    [`&:hover .${showOnHoverClass}`]: {
      opacity: 1,
    },
  },
  copyButton: {
    backgroundColor: theme.palette.common.white,
    border: `1px solid ${theme.palette.primary.main}`,
    borderRadius: theme.spacing(1),
    color: theme.palette.text.secondary,
    height: theme.spacing(4),
    minWidth: 0,
    padding: 0,
    position: 'absolute',
    right: theme.spacing(1),
    top: theme.spacing(1),
    width: theme.spacing(4),
    '&:hover': {
      backgroundColor: listhoverColor,
    },
  },
  errorMessage: {
    fontFamily: 'monospace',
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-all',
    wordWrap: 'break-word',
  },
  expandButton: {
    backgroundColor: theme.palette.text.secondary,
    borderRadius: 16,
    bottom: theme.spacing(4),
    color: theme.palette.common.white,
    fontFamily: bodyFontFamily,
    fontSize: smallFontSize,
    fontWeight: 'bold',
    left: '50%',
    padding: `${theme.spacing(1)}px ${theme.spacing(3)}px`,
    position: 'absolute',
    transform: 'translateX(-50%)',
    '&:hover': {
      backgroundColor: theme.palette.text.primary,
    },
    [`&.${expandedClass}`]: {
      bottom: theme.spacing(2),
    },
  },
  // Apply this class to an ancestor container to have the expandable text
  // use an alternate background color scheme.
  nestedParent: {
    backgroundColor: nestedListColor,
  },
}));

export interface ExpandableMonospaceTextProps {
  initialExpansionState?: boolean;
  text: string;
  onExpandCollapse?(expanded: boolean): void;
}

/** An expandable/collapsible container which renders the provided text in a
 * monospace font. It also provides a button to copy the text.
 */
export const ExpandableMonospaceText: React.FC<ExpandableMonospaceTextProps> = ({
  onExpandCollapse,
  initialExpansionState = false,
  text,
}) => {
  const [expanded, setExpanded] = React.useState(initialExpansionState);
  const styles = useExpandableMonospaceTextStyles();
  const onClickExpand = () => {
    setExpanded(!expanded);
    typeof onExpandCollapse === 'function' && onExpandCollapse(!expanded);
  };
  const onClickCopy = () => copyToClipboard(text);

  const expandButtonText = expanded ? 'Collapse' : 'Click to expand inline';

  return (
    <div
      className={classnames(styles.container, {
        [expandedClass]: expanded,
      })}
    >
      <div className={styles.errorMessage}>{`${text}`}</div>
      {expanded ? null : <div className={styles.bottomFade} />}
      <div className={classnames(styles.actionContainer, showOnHoverClass)}>
        <ButtonBase
          disableRipple={true}
          onClick={onClickExpand}
          className={classnames(styles.expandButton, showOnHoverClass, {
            [expandedClass]: expanded,
          })}
        >
          {expandButtonText}
        </ButtonBase>
        <Button
          color="primary"
          className={styles.copyButton}
          onClick={onClickCopy}
          variant="outlined"
        >
          <FileCopy fontSize="small" />
        </Button>
      </div>
    </div>
  );
};
