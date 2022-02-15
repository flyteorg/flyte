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
const HTMLExternalsWebpackPlugin = require('html-webpack-externals-plugin');
const nodeExternals = require('webpack-node-externals');
const UglifyJsPlugin = require('uglifyjs-webpack-plugin');
import { processEnv as env } from './env';

const packageJson: {
  dependencies: { [libName: string]: string };
  devDependencies: { [libName: string]: string };
} = require(require.resolve('./package.json'));

/** Current service name */
export const serviceName = process.env.SERVICE_NAME || 'not set';

/** Absolute path to webpack output folder */
export const dist = path.join(__dirname, 'dist');

/** Webpack public path. All emitted assets will have relative path to this path */
export const publicPath = `${env.BASE_URL}/assets/`;

/** True if we are in development mode */
export const isDev = env.NODE_ENV === 'development';

/** True if we are in production mode */
export const isProd = env.NODE_ENV === 'production';

/** CSS module class name pattern */
export const localIdentName = isDev ? '[local]_[hash:base64:3]' : '[hash:base64:6]';

/** Sourcemap configuration */
const devtool = isDev ? 'cheap-source-map' : undefined;

// Report current configuration
console.log(chalk.cyan('Exporting Webpack config with following configurations:'));
console.log(chalk.blue('Environment:'), chalk.green(env.NODE_ENV));
console.log(chalk.blue('Output directory:'), chalk.green(path.resolve(dist)));
console.log(chalk.blue('Public path:'), chalk.green(publicPath));

/** SourceMap loader allows Source Map support in JavaScript files. */
export const sourceMapLoader: webpack.Loader = 'source-map-loader';

/** Load static files like images */
export const fileLoader: webpack.Loader = {
  loader: 'file-loader',
  options: {
    publicPath,
    name: '[hash:8].[ext]'
  }
};

/** Common webpack resolve options */
export const resolve: webpack.Resolve = {
  /** Base directories that Webpack will look to resolve absolutely imported modules */
  modules: ['src', 'node_modules'],
  /** Extension that are allowed to be omitted from import statements */
  extensions: ['.ts', '.tsx', '.js', '.jsx'],
  /** "main" fields in package.json files to resolve a CommonJS module for */
  mainFields: ['browser', 'module', 'main']
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

/** Map of libraries that are loaded from CDN and are omitted from emitted JS */
export const clientExternals: webpack.ExternalsElement = {
  react: 'React',
  'react-dom': 'ReactDOM'
};

/** A map of library names to their CDN values */
export const clientExternalsCDNUrls: { [libName: string]: string } = {
  react: `https://unpkg.com/react@${absoluteVersion(packageJson.devDependencies.react)}/umd/react.production.min.js`,
  'react-dom': `https://unpkg.com/react-dom@${absoluteVersion(
    packageJson.devDependencies['react-dom']
  )}/umd/react-dom.production.min.js`
};

/** Minification options for HTMLWebpackPlugin */
export const htmlMinifyConfig: HTMLWebpackPlugin.MinifyConfig = {
  minifyCSS: true,
  minifyJS: false,
  removeComments: true,
  collapseInlineTagWhitespace: true,
  collapseWhitespace: true
};

/** FavIconWebpackPlugin icon options */
export const favIconIconsOptions = {
  persistentCache: isDev,
  android: false,
  appleIcon: false,
  appleStartup: false,
  coast: false,
  favicons: true,
  firefox: false,
  opengraph: false,
  twitter: false,
  yandex: false,
  windows: false
};

/** Adds sourcemap support */
export const sourceMapRule: webpack.Rule = {
  test: /\.js$/,
  enforce: 'pre',
  use: sourceMapLoader
};

/** Rule for images, icons and fonts */
export const imageAndFontsRule: webpack.Rule = {
  test: /\.(ico|jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2)(\?.*)?$/,
  use: fileLoader
};

/** Generates HTML file that includes webpack assets */
export const htmlPlugin = new HTMLWebpackPlugin({
  template: './src/assets/index.html',
  inject: 'body',
  minify: isProd ? htmlMinifyConfig : false,
  hash: false,
  showErrors: isDev
});

/** Generate all favorite icons based on one icon file */
export const favIconPlugin = new FavIconWebpackPlugin({
  logo: path.resolve(__dirname, 'src/assets/favicon.png'),
  prefix: '[hash:8]/',
  persistentCache: false,
  icons: favIconIconsOptions
});

/** Write client stats to a JSON file for production */
export const statsWriterPlugin = new StatsWriterPlugin({
  filename: 'client-stats.json',
  fields: ['chunks', 'publicPath', 'assets', 'assetsByChunkName', 'assetsByChunkId']
});

/** Inject CDN URLs for externals */
export const htmlExternalsPlugin = new HTMLExternalsWebpackPlugin({
  externals: Object.keys(clientExternals).map(libName => ({
    module: libName,
    entry: clientExternalsCDNUrls[libName],
    global: clientExternals[libName]
  }))
});

/** Gzip assets */
export const compressionPlugin = new CompressionWebpackPlugin({
  algorithm: 'gzip',
  test: /\.(js|css|html)$/,
  threshold: 10240,
  minRatio: 0.8
});

/** Define "process.env" in client app. Only provide things that can be public */
export const getDefinePlugin = (isServer: boolean) =>
  new webpack.DefinePlugin({
    'process.env': isServer
      ? 'process.env'
      : Object.keys(env).reduce(
          (result, key: string) => ({
            ...result,
            [key]: JSON.stringify((env as any)[key])
          }),
          {}
        ),
    __isServer: isServer
  });

/** Enables Webpack HMR */
export const hmrPlugin = new webpack.HotModuleReplacementPlugin();

/** Do not emit anything if there is an error */
export const noErrorsPlugin = new webpack.NoEmitOnErrorsPlugin();

/** Limit server chunks to be only one. No need to split code in server */
export const limitChunksPlugin = new webpack.optimize.LimitChunkCountPlugin({
  maxChunks: 1
});

const typescriptRule = {
  test: /\.tsx?$/,
  exclude: /node_modules/,
  use: ['babel-loader', { loader: 'ts-loader', options: { transpileOnly: true } }]
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
  externals: isProd ? clientExternals : undefined,
  module: {
    rules: [sourceMapRule, typescriptRule, imageAndFontsRule]
  },
  optimization: {
    noEmitOnErrors: isDev,
    minimizer: [
      new UglifyJsPlugin({
        parallel: true,
        sourceMap: true,
        uglifyOptions: {
          compress: {
            collapse_vars: false,
            warnings: false
          },
          output: { comments: false }
        }
      })
    ],
    splitChunks: {
      cacheGroups: {
        vendor: {
          chunks: 'initial',
          enforce: true,
          name: 'vendor',
          priority: 10,
          test: /node_modules/
        }
      }
    }
  },
  output: {
    publicPath,
    path: dist,
    filename: '[name]-[hash:8].js',
    chunkFilename: '[name]-[chunkhash].chunk.js',
    crossOriginLoading: 'anonymous'
  },
  get plugins() {
    const plugins: webpack.Plugin[] = [
      htmlPlugin,
      new ForkTsCheckerWebpackPlugin({ checkSyntacticErrors: true }),
      favIconPlugin,
      statsWriterPlugin,
      getDefinePlugin(false),
      new ServePlugin({
        middleware: (app, builtins) =>
          app.use(async (ctx, next) => {
            ctx.setHeader('Content-Type', 'application/javascript; charset=UTF-8');
            await next();
          }),
        port: 7777
      })
    ];

    // Apply production specific configs
    if (isProd) {
      // plugins.push(uglifyPlugin);
      plugins.push(htmlExternalsPlugin);
      plugins.push(compressionPlugin);
    }

    if (isDev) {
      return [
        ...plugins
        // namedModulesPlugin,
      ];
    }

    return plugins;
  }
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
  devtool: isProd ? devtool : undefined,
  entry: ['babel-polyfill', './src/server'],
  module: {
    rules: [sourceMapRule, typescriptRule, imageAndFontsRule]
  },
  externals: [nodeExternals({ whitelist: /lyft/ })],
  output: {
    path: dist,
    filename: 'server.js',
    libraryTarget: 'commonjs2'
  },
  plugins: [
    limitChunksPlugin,
    new ForkTsCheckerWebpackPlugin({ checkSyntacticErrors: true }),
    getDefinePlugin(true)
    // namedModulesPlugin,
  ]
};

export default [clientConfig, serverConfig];
