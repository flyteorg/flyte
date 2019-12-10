import { CssBaseline } from '@material-ui/core';
import { ThemeProvider } from '@material-ui/styles';
import { env } from 'common/env';
import { debug, debugPrefix } from 'common/log';
import { APIContext, useAPIState } from 'components/data/apiContext';
import { LoginExpiredHandler } from 'components/Errors/LoginExpiredHandler';
import { skeletonColor, skeletonHighlightColor } from 'components/Theme';
import { muiTheme } from 'components/Theme/muiTheme';
import * as React from 'react';
import { Helmet } from 'react-helmet';
import { hot } from 'react-hot-loader';
import { SkeletonTheme } from 'react-loading-skeleton';
import { Router } from 'react-router-dom';
import { ApplicationRouter } from 'routes/ApplicationRouter';
import { history } from 'routes/history';
import { NavBarRouter } from 'routes/NavBarRouter';
import { ErrorBoundary } from '../common';

export const AppComponent: React.StatelessComponent<{}> = () => {
    if (env.NODE_ENV === 'development') {
        debug.enable(`${debugPrefix}*:*`);
    }
    const apiState = useAPIState();

    return (
        <ThemeProvider theme={muiTheme}>
            <APIContext.Provider value={apiState}>
                <SkeletonTheme
                    color={skeletonColor}
                    highlightColor={skeletonHighlightColor}
                >
                    <CssBaseline />
                    <Helmet>
                        <title>Flyte Console</title>
                        <meta name="viewport" content="width=device-width" />
                    </Helmet>
                    <Router history={history}>
                        <ErrorBoundary fixed={true}>
                            <NavBarRouter />
                            <ApplicationRouter />
                            <LoginExpiredHandler />
                        </ErrorBoundary>
                    </Router>
                </SkeletonTheme>
            </APIContext.Provider>
        </ThemeProvider>
    );
};

export const App = hot(module)(AppComponent);
