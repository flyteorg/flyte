import { env } from 'common/env';
import 'intersection-observer';
import 'protobuf';
import * as React from 'react';
import * as ReactDOM from 'react-dom';

const render = (Component: React.FC) => {
    ReactDOM.render(<Component />, document.getElementById('react-app'));
};

const initializeApp = () => {
    const App = require('./components/App/App').App;

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
