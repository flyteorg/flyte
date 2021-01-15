const webpack = require('webpack');

const env = require('../env').processEnv;

/** Adds sourcemap support */
const sourceMapRule = {
    test: /\.js$/,
    enforce: 'pre',
    use: 'source-map-loader'
};

/** Define "process.env" in client app. Only provide things that can be public */
const definePlugin = new webpack.DefinePlugin({
    'window.env': Object.keys(env).reduce(
        (result, key) =>
            Object.assign(result, {
                [key]: JSON.stringify(env[key])
            }),
        {}
    ),
    __isServer: false
});

module.exports = ({ config: defaultConfig }) => {
    // Insert typescript loader ahead of others
    defaultConfig.module.rules.unshift(sourceMapRule, {
        test: /\.(ts|tsx)$/,
        loader: ['ts-loader']
    });

    defaultConfig.resolve.extensions.push('.ts', '.tsx');
    defaultConfig.plugins.push(definePlugin);

    defaultConfig.resolve.modules = ['src', 'node_modules'];
    defaultConfig.resolve.mainFields = ['browser', 'module', 'main'];

    return defaultConfig;
};
