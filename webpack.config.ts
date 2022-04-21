// tslint:disable:no-var-requires
// tslint:disable:no-console
import chalk from 'chalk';
import * as CompressionWebpackPlugin from 'compression-webpack-plugin';
import * as HTMLWebpackPlugin from 'html-webpack-plugin';
import * as path from 'path';
import * as webpack from 'webpack';

const { WebpackPluginServe: ServePlugin } = require('webpack-plugin-serve');
const { StatsWriterPlugin } = require('webpack-stats-plugin');
const FavIconWebpackPlugin = require('favicons-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');
const nodeExternals = require('webpack-node-externals');

import { processEnv as env } from './env';

const packageJson: {
  dependencies: { [libName: string]: string };
  devDependencies: { [libName: string]: string };
} = require(require.resolve('./package.json'));

/** Current service name */
export const serviceName = process.env.SERVICE_NAME || 'not set';

/** Absolute path to webpack output folder */
export const dist = path.join(__dirname, 'dist');

/** Webpack public path. All emitted assets will have relative path to this path
 * every time it is changed - the index.js app.use should also be updated.
 */
export const publicPath = `${env.BASE_URL}/assets/`;

/** True if we are in development mode */
export const isDev = env.NODE_ENV === 'development';

/** True if we are in production mode */
export const isProd = env.NODE_ENV === 'production';

/** CSS module class name pattern */
export const localIdentName = isDev ? '[local]_[fullhash:base64:3]' : '[fullhash:base64:6]';

/** Sourcemap configuration */
const devtool = isDev ? 'cheap-source-map' : undefined;

// Report current configuration
console.log(chalk.cyan('Exporting Webpack config with following configurations:'));
console.log(chalk.blue('Environment:'), chalk.green(env.NODE_ENV));
console.log(chalk.blue('Output directory:'), chalk.green(path.resolve(dist)));
console.log(chalk.blue('Public path:'), chalk.green(publicPath));

/** Common webpack resolve options */
export const resolve: webpack.ResolveOptions = {
  /** Base directories that Webpack will look to resolve absolutely imported modules */
  modules: ['src', 'node_modules'],
  /** Extension that are allowed to be omitted from import statements */
  extensions: ['.ts', '.tsx', '.js', '.jsx'],
  /** "main" fields in package.json files to resolve a CommonJS module for */
  mainFields: ['browser', 'module', 'main'],
};

/** Get clean version of a version string of package.json entry for a package by
 * extracting only alphanumerics, hyphen, and period. Note that this won't
 * produce a valid URL for all possible NPM version strings, but should be fine
 * on those that are absolute version references.
 * Examples: '1', '1.0', '1.2.3', '1.2.3-alpha.0'
 */
export function absoluteVersion(version: string) {
  return version.replace(/[^\d.\-a-z]/g, '');
}

/** CDN path in case we would use minified react and react-DOM */
const cdnReact = `https://unpkg.com/react@${absoluteVersion(
  packageJson.devDependencies.react,
)}/umd/react.production.min.js`;
const cdnReactDOM = `https://unpkg.com/react-dom@${absoluteVersion(
  packageJson.devDependencies['react-dom'],
)}/umd/react-dom.production.min.js`;

/** Minification options for HTMLWebpackPlugin */
export const htmlMinifyConfig: HTMLWebpackPlugin.MinifyOptions = {
  minifyCSS: true,
  minifyJS: false,
  removeComments: true,
  collapseInlineTagWhitespace: true,
  collapseWhitespace: true,
};

/** Adds sourcemap support */
export const sourceMapRule: webpack.RuleSetRule = {
  test: /\.js$/,
  enforce: 'pre',
  use: ['source-map-loader'],
};

/** Rule for images, icons and fonts */
export const imageAndFontsRule: webpack.RuleSetRule = {
  test: /\.(ico|jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2)(\?.*)?$/,
  type: 'asset/resource',
};

/** Generates HTML file that includes webpack assets */
export const htmlPlugin = new HTMLWebpackPlugin({
  template: './src/assets/index.html',
  inject: 'body',
  minify: isProd ? htmlMinifyConfig : false,
  hash: false,
  showErrors: isDev,
});

export const favIconPlugin = new FavIconWebpackPlugin({
  logo: path.resolve(__dirname, 'src/assets/favicon.png'),
  // we can add '[fullhash:8]/' to the end of the file in future
  // if this one will be changed - ensure that OneClick will still be working
  prefix: './',
});

/** Write client stats to a JSON file for production */
export const statsWriterPlugin = new StatsWriterPlugin({
  filename: 'client-stats.json',
  fields: ['chunks', 'publicPath', 'assets', 'assetsByChunkName', 'assetsByChunkId'],
});

/** Gzip assets */
export const compressionPlugin = new CompressionWebpackPlugin({
  algorithm: 'gzip',
  test: /\.(js|css|html)$/,
  threshold: 10240,
  minRatio: 0.8,
});

/** Define "process.env" in client app. Only provide things that can be public */
export const getDefinePlugin = (isServer: boolean) =>
  new webpack.DefinePlugin({
    'process.env': isServer
      ? 'process.env'
      : Object.keys(env).reduce(
          (result, key: string) => ({
            ...result,
            [key]: JSON.stringify((env as any)[key]),
          }),
          {},
        ),
    __isServer: isServer,
  });

/** Enables Webpack HMR */
export const hmrPlugin = new webpack.HotModuleReplacementPlugin();

/** Limit server chunks to be only one. No need to split code in server */
export const limitChunksPlugin = new webpack.optimize.LimitChunkCountPlugin({
  maxChunks: 1,
});

const typescriptRule = {
  test: /\.tsx?$/,
  exclude: /node_modules/,
  use: ['babel-loader', { loader: 'ts-loader', options: { transpileOnly: true } }],
};

/**
 * Client configuration
 *
 * Client is compiled into multiple chunks that are result to dynamic imports.
 */
export const clientConfig: webpack.Configuration = {
  devtool,
  resolve,
  name: 'client',
  target: 'web',
  get entry() {
    const entry = ['babel-polyfill', './src/client'];

    if (isDev) {
      return ['webpack-hot-middleware/client', ...entry];
    }

    return entry;
  },
  mode: isDev ? 'development' : 'production',
  module: {
    rules: [sourceMapRule, typescriptRule, imageAndFontsRule],
  },
  optimization: {
    emitOnErrors: isProd,
    splitChunks: {
      cacheGroups: {
        vendor: {
          chunks: 'initial',
          enforce: true,
          name: 'vendor',
          priority: 10,
          test: /[\\/]node_modules/,
        },
      },
    },
  },
  output: {
    publicPath,
    path: dist,
    filename: '[name]-[fullhash:8].js',
    chunkFilename: '[name]-[chunkhash].chunk.js',
    crossOriginLoading: 'anonymous',
  },
  get plugins() {
    const plugins: webpack.WebpackPluginInstance[] = [
      htmlPlugin,
      new ForkTsCheckerWebpackPlugin(),
      favIconPlugin,
      statsWriterPlugin,
      getDefinePlugin(false),
      new ServePlugin({
        middleware: (app, builtins) =>
          app.use(async (ctx, next) => {
            ctx.setHeader('Content-Type', 'application/javascript; charset=UTF-8');
            await next();
          }),
        port: 7777,
      }),
    ];

    // Apply production specific configs
    if (isProd) {
      plugins.push(compressionPlugin);
    }

    return plugins;
  },
};

/**
 * Server configuration
 *
 * Server bundle is compiled as a CommonJS package that exports an Express middleware
 */
export const serverConfig: webpack.Configuration = {
  resolve,
  name: 'server',
  target: 'node',
  mode: isDev ? 'development' : 'production',
  devtool: isProd ? devtool : undefined,
  entry: ['babel-polyfill', './src/server'],
  module: {
    rules: [sourceMapRule, typescriptRule, imageAndFontsRule],
  },
  externals: [nodeExternals({ allowlist: /lyft/ })],
  output: {
    path: dist,
    filename: 'server.js',
    libraryTarget: 'commonjs2',
    // assetModuleFilename: `${publicPath}/[fullhash:8].[ext]`
  },
  plugins: [limitChunksPlugin, new ForkTsCheckerWebpackPlugin(), getDefinePlugin(true)],
};

export default [clientConfig, serverConfig];
