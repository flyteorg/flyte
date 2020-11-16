import { Theme } from '@material-ui/core';
import { makeStyles } from '@material-ui/styles';
import { smallFontSize } from 'components/Theme';

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
        padding: theme.spacing(2)
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
    }
}));
