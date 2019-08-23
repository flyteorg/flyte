import { CssBaseline } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ThemeProvider } from '@material-ui/styles';
import * as React from 'react';
import { ErrorBoundary } from '../src/components/common';
import { muiTheme } from '../src/components/Theme/muiTheme';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        padding: theme.spacing(1)
    }
}));

export const StorybookContainer: React.StatelessComponent = ({ children }) => (
    <ThemeProvider theme={muiTheme}>
        <CssBaseline />
        <ErrorBoundary>
            <div className={useStyles().container}>{children}</div>
        </ErrorBoundary>
    </ThemeProvider>
);
