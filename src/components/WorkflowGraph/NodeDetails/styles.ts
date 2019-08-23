import { makeStyles, Theme } from '@material-ui/core/styles';

export const useStyles = makeStyles((theme: Theme) => {
    const padding = `${theme.spacing(2)}px`;
    return {
        container: {
            display: 'flex',
            flexDirection: 'column',
            height: '100%'
        },
        content: {
            overflowY: 'auto'
        },
        header: {
            borderBottom: `${theme.spacing(1)}px solid ${theme.palette.divider}`
        },
        headerContent: {
            padding: `0 ${padding} ${padding} ${padding}`
        },
        tabs: {
            borderBottom: `1px solid ${theme.palette.divider}`
        }
    };
});
