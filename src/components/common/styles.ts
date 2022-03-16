import { makeStyles, Theme } from '@material-ui/core/styles';
import {
  buttonHoverColor,
  dangerousButtonBorderColor,
  dangerousButtonColor,
  dangerousButtonHoverColor,
  mutedPrimaryTextColor,
  smallFontSize,
} from 'components/Theme/constants';

const unstyledLinkProps = {
  textDecoration: 'none',
  color: 'inherit',
};

export const horizontalListSpacing = (spacing: number) => ({
  '& > :not(:first-child)': {
    marginLeft: spacing,
  },
});

export const buttonGroup = (theme: Theme) => ({
  ...horizontalListSpacing(theme.spacing(1)),
  display: 'flex',
});

export const useCommonStyles = makeStyles((theme: Theme) => ({
  buttonWhiteOutlined: {
    backgroundColor: theme.palette.common.white,
    '&:hover': {
      backgroundColor: buttonHoverColor,
    },
  },
  codeBlock: {
    display: 'block',
    fontFamily: 'monospace',
    marginTop: theme.spacing(1),
    textTransform: 'none',
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-all',
    wordWrap: 'break-word',
  },
  dangerousButton: {
    borderColor: dangerousButtonBorderColor,
    color: dangerousButtonColor,
    '&:hover': {
      borderColor: dangerousButtonHoverColor,
      color: dangerousButtonHoverColor,
    },
  },
  detailsPanelCard: {
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  detailsPanelCardContent: {
    padding: `${theme.spacing(2)}px ${theme.spacing(3)}px`,
  },
  errorText: {
    color: theme.palette.error.main,
  },
  flexCenter: {
    alignItems: 'center',
    display: 'flex',
  },
  flexFill: {
    flex: '1 1 100%',
  },
  formButtonGroup: {
    ...buttonGroup(theme),
    justifyContent: 'center',
    marginTop: theme.spacing(1),
  },
  formControlLabelSmall: {
    // This is adjusting the negative margin used by
    // FormControlLabel to make sure the control lines up correctly
    marginLeft: -3,
    marginTop: theme.spacing(1),
  },
  formControlSmall: {
    marginRight: theme.spacing(1),
    padding: 0,
  },
  hintText: {
    color: theme.palette.text.hint,
  },
  iconLeft: {
    marginRight: theme.spacing(1),
  },
  iconRight: {
    marginLeft: theme.spacing(1),
  },
  iconSecondary: {
    color: theme.palette.text.secondary,
  },
  linkUnstyled: {
    ...unstyledLinkProps,
    '&:hover': {
      ...unstyledLinkProps,
    },
  },
  listUnstyled: {
    listStyle: 'none',
    margin: 0,
    padding: 0,
    '& li': {
      padding: 0,
    },
  },
  microHeader: {
    textTransform: 'uppercase',
    fontWeight: 'bold',
    fontSize: smallFontSize,
    lineHeight: '.9375rem',
  },
  mutedHeader: {
    color: mutedPrimaryTextColor,
    fontWeight: 'bold',
    fontSize: '1rem',
    lineHeight: '1.375rem',
  },
  primaryLink: {
    color: theme.palette.primary.main,
    cursor: 'pointer',
    fontWeight: 'bold',
    textDecoration: 'none',
    '&:hover': {
      color: theme.palette.primary.main,
      textDecoration: 'underline',
    },
  },
  secondaryLink: {
    color: theme.palette.text.secondary,
    cursor: 'pointer',
    fontWeight: 'bold',
    textDecoration: 'none',
    '&:hover': {
      color: theme.palette.text.secondary,
      textDecoration: 'underline',
    },
  },
  smallDropdownWindow: {
    border: `1px solid ${theme.palette.divider}`,
    marginTop: theme.spacing(0.5),
  },
  textMonospace: {
    textTransform: 'none',
    fontFamily: 'monospace',
  },
  textMuted: {
    color: theme.palette.grey[500],
  },
  textSmall: {
    fontSize: smallFontSize,
    lineHeight: 1.25,
  },
  textWrapped: {
    overflowWrap: 'break-word',
    wordBreak: 'break-all',
  },
  truncateText: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
}));
