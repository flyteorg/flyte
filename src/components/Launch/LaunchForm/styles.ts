import { Theme } from '@material-ui/core';
import { makeStyles } from '@material-ui/styles';
import {
    interactiveTextColor,
    smallFontSize
} from 'components/Theme/constants';

export const useStyles = makeStyles((theme: Theme) => ({
    footer: {
        padding: theme.spacing(2)
    },
    formControl: {
        padding: `${theme.spacing(1.5)}px 0`
    },
    header: {
        padding: theme.spacing(2),
        width: '100%'
    },
    inputsSection: {
        padding: theme.spacing(2),
        maxHeight: theme.spacing(90)
    },
    inputLabel: {
        color: theme.palette.text.hint,
        fontSize: smallFontSize
    },
    root: {
        display: 'flex',
        flexDirection: 'column',
        width: '100%'
    },
    sectionHeader: {
        marginBottom: theme.spacing(1),
        marginTop: theme.spacing(1)
    },
    advancedOptions: {
        color: interactiveTextColor,
        justifyContent: 'flex-end'
    },
    noBorder: {
        '&:before': {
            height: 0
        }
    },
    summaryWrapper: {
        padding: 0
    },
    detailsWrapper: {
        paddingLeft: 0,
        paddingRight: 0,
        flexDirection: 'column',
        '& section': {
            flex: 1
        }
    },
    collapsibleSection: {
        margin: '0 -16px'
    }
}));
