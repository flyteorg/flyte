import { CssBaseline } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ThemeProvider } from '@material-ui/styles';
import * as React from 'react';
import { QueryClientProvider } from 'react-query';
import { ErrorBoundary } from '../src/components/common';
import { createQueryClient } from '../src/components/data/queryCache';
import { muiTheme } from '../src/components/Theme/muiTheme';

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        display: 'flex',
        padding: theme.spacing(1)
    }
}));

export const StorybookContainer: React.FC = ({ children }) => {
    const [queryClient] = React.useState(() => createQueryClient());
    return (
        <ThemeProvider theme={muiTheme}>
            <CssBaseline />
            <ErrorBoundary>
                <QueryClientProvider client={queryClient}>
                <div className={useStyles().container}>{children}</div>
                </QueryClientProvider>
            </ErrorBoundary>
        </ThemeProvider>
    );
};
