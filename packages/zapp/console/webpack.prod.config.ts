import * as webpack from 'webpack';
import * as CompressionWebpackPlugin from 'compression-webpack-plugin';
import * as HTMLWebpackPlugin from 'html-webpack-plugin';

const { merge } = require('webpack-merge');

const common = require('./webpack.common.config');

const packageJson: {
  dependencies: { [libName: string]: string };
  devDependencies: { [libName: string]: string };
} = require(require.resolve('./package.json'));

/** Get clean version of a version string of package.json entry for a package by
 * extracting only alphanumerics, hyphen, and period. Note that this won't
 * produce a valid URL for all possible NPM version strings, but should be fine
 * on those that are absolute version references.
 * Examples: '1', '1.0', '1.2.3', '1.2.3-alpha.0'
 */
function absoluteVersion(version: string) {
  return version.replace(/[^\d.\-a-z]/g, '');
}

/** CDN path to use minified react and react-DOM instead of prebuild version */
const cdnReact = `https://unpkg.com/react@${absoluteVersion(
  packageJson.devDependencies.react,
)}/umd/react.production.min.js`;
const cdnReactDOM = `https://unpkg.com/react-dom@${absoluteVersion(
  packageJson.devDependencies['react-dom'],
)}/umd/react-dom.production.min.js`;

/**
 * Client configuration
 *
 * Client is compiled into multiple chunks that are result to dynamic imports.
 */
export const clientConfig: webpack.Configuration = merge(common.default.clientConfig, {
  mode: 'production',
  optimization: {
    emitOnErrors: true,
  },
  plugins: [
    new CompressionWebpackPlugin({
      algorithm: 'gzip',
      test: /\.(js|css|html)$/,
      threshold: 10240,
      minRatio: 0.8,
    }),
    new HTMLWebpackPlugin({
      template: './src/assets/index.html',
      inject: 'body',
      minify: {
        minifyCSS: true,
        minifyJS: false,
        removeComments: true,
        collapseInlineTagWhitespace: true,
        collapseWhitespace: true,
      },
      hash: false,
      showErrors: false,
      cdnReact,
      cdnReactDOM,
    }),
  ],
  externals: {
    'react-dom': 'ReactDOM',
    react: 'React',
  },
});

/**
 * Server configuration
 *
 * Server bundle is compiled as a CommonJS package that exports an Express middleware
 */
export const serverConfig: webpack.Configuration = merge(common.default.serverConfig, {
  mode: 'production',
});

export default [clientConfig, serverConfig];
