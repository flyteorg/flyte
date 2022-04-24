import * as React from 'react';
import { CssBaseline } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ThemeProvider } from '@material-ui/styles';
import { QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router-dom';
import { SnackbarProvider } from 'notistack';

import { ErrorBoundary } from '../packages/zapp/console/src/components/common/ErrorBoundary';
import { createQueryClient } from '../packages/zapp/console/src/components/data/queryCache';
import { muiTheme } from '../packages/zapp/console/src/components/Theme/muiTheme';
import { LocalCacheProvider } from '../packages/zapp/console/src/basics/LocalCache/ContextProvider';
import { FeatureFlagsProvider } from '../packages/zapp/console/src/basics/FeatureFlags';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    display: 'flex',
    padding: theme.spacing(1),
  },
}));

export const StorybookContainer: React.FC = ({ children }) => {
  const [queryClient] = React.useState(() => createQueryClient());
  return (
    <ThemeProvider theme={muiTheme}>
      <CssBaseline />
      <ErrorBoundary>
        <MemoryRouter>
          <FeatureFlagsProvider>
            <LocalCacheProvider>
              <SnackbarProvider
                maxSnack={2}
                anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
              >
                <QueryClientProvider client={queryClient}>
                  <div className={useStyles().container}>{children}</div>
                </QueryClientProvider>
              </SnackbarProvider>
            </LocalCacheProvider>
          </FeatureFlagsProvider>
        </MemoryRouter>
      </ErrorBoundary>
    </ThemeProvider>
  );
};
