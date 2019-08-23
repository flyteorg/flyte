import 'protobuf';

import * as React from 'react';
import * as ReactDOM from 'react-dom';

import { env } from 'common/env';

const render = (Component: React.StatelessComponent) => {
    ReactDOM.render(<Component />, document.getElementById('react-app'));
};

const initializeApp = () => {
    // tslint:disable-next-line:no-var-requires
    const App = require('./components/App').App;

    if (env.NODE_ENV === 'development') {
        // We use style-loader in dev mode, but it causes a FOUC and some initial styling issues
        // so we'll give it time to add the styles before initial render.
        setTimeout(() => render(App), 500);
    } else {
        render(App);
    }
};

if (document.body) {
    initializeApp();
} else {
    window.addEventListener('DOMContentLoaded', initializeApp, false);
}
