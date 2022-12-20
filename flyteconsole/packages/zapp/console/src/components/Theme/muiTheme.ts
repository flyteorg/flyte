import { createMuiTheme } from '@material-ui/core/styles';
import {
  bodyFontFamily,
  headerFontFamily,
  inputFocusBorderColor,
  mutedButtonColor,
  primaryColor,
  primaryDarkColor,
  primaryLightColor,
  primaryTextColor,
  secondaryColor,
  secondaryTextColor,
  selectedActionColor,
  whiteColor,
} from './constants';
import { COLOR_SPECTRUM } from './colorSpectrum';

const theme = createMuiTheme({
  palette: {
    primary: {
      main: primaryColor,
      light: primaryLightColor,
      dark: primaryDarkColor,
    },
    secondary: {
      main: secondaryColor,
    },
    background: {
      default: whiteColor,
      // default: '#F4F4FA',
    },
    text: {
      primary: primaryTextColor,
      secondary: secondaryTextColor,
    },
    action: {
      selected: selectedActionColor,
      hoverOpacity: 0.15,
    },
  },
  props: {
    MuiIconButton: {
      centerRipple: false,
    },
    MuiPopover: {
      elevation: 7,
    },
    MuiAppBar: {
      elevation: 1,
      color: 'secondary',
    },
    MuiTabs: {
      indicatorColor: 'primary',
    },
    MuiPaper: {
      elevation: 0,
    },
  },
  shape: {
    borderRadius: 8,
  },
  typography: {
    // https://material-ui.com/style/typography/
    fontFamily: headerFontFamily,
    body1: {
      fontSize: '14px',
      fontFamily: bodyFontFamily,
    },
    body2: {
      fontSize: '14px',
      fontFamily: bodyFontFamily,
    },
    h3: {
      fontSize: '16px',
      fontWeight: 'bold',
      lineHeight: '22px',
    },
    h4: {
      fontSize: '14px',
      fontWeight: 'bold',
      lineHeight: '20px',
    },
    h6: {
      fontSize: '.875rem',
      fontWeight: 'bold',
      lineHeight: 1.357,
    },
    overline: {
      fontFamily: headerFontFamily,
      fontSize: '.75rem',
      fontWeight: 'bold',
      letterSpacing: 'normal',
      lineHeight: '1rem',
      textTransform: 'uppercase',
    },
    subtitle1: {
      fontFamily: headerFontFamily,
      fontSize: '.75rem',
      letterSpacing: 'normal',
      lineHeight: '1rem',
    },
  },
});

export const muiTheme = {
  ...theme,
  overrides: {
    MuiButton: {
      root: {
        fontWeight: 600,
      },
      label: {
        textTransform: 'initial',
      },
      contained: {
        boxShadow: 'none',
        color: whiteColor,
        ['&:active']: {
          boxShadow: 'none',
        },
      },
      outlined: {
        backgroundColor: 'transparent',
      },
      outlinedPrimary: {
        border: `1px solid ${theme.palette.primary.main}`,
        color: theme.palette.primary.main,
        ['&:hover']: {
          borderColor: theme.palette.primary.dark,
          color: theme.palette.primary.dark,
        },
      },
      text: {
        color: theme.palette.text.secondary,
        ['&:hover']: {
          color: theme.palette.text.primary,
        },
      },
    },
    MuiOutlinedInput: {
      root: {
        '& $notchedOutline': {
          borderRadius: 4,
        },
        '&$focused $notchedOutline': {
          borderColor: inputFocusBorderColor,
        },
      },
    },
    MuiTab: {
      labelIcon: {
        minHeight: '64px',
        paddingTop: 0,
      },
      wrapper: {
        flexDirection: 'row',
        ['& svg']: {
          marginLeft: 12,
          position: 'relative',
          left: 12,
        },
      },
      root: {
        fontWeight: 600,
        fontSize: '.875rem',
        margin: `0 ${theme.spacing(1.75)}px`,
        minWidth: 0,
        padding: `0 ${theme.spacing(1.75)}px`,
        textTransform: 'none',
        [theme.breakpoints.up('md')]: {
          fontSize: '.875rem',
          margin: `0 ${theme.spacing(1.75)}px`,
          minWidth: 0,
          padding: `0 ${theme.spacing(1.75)}px`,
        },
        ['&:hover']: {
          backgroundColor: theme.palette.action.hover,
        },
      },
    },
    MuiTabs: {
      indicator: {
        height: 4,
      },
    },
    MuiSwitch: {
      switchBase: {
        // Controls default (unchecked) color for the thumb
        color: whiteColor,
      },
      colorSecondary: {
        '&$checked': {
          // Controls checked color for the thumb
          color: whiteColor,
        },
      },
      track: {
        // Controls default (unchecked) color for the track
        opacity: 1,
        backgroundColor: mutedButtonColor,
        '$checked$checked + &': {
          // Controls checked color for the track
          opacity: 1,
          backgroundColor: COLOR_SPECTRUM.mint20.color,
        },
      },
    },
  },
};
