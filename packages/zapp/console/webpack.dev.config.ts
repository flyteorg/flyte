import * as webpack from 'webpack';
import * as HTMLWebpackPlugin from 'html-webpack-plugin';
import * as path from 'path';
import chalk from 'chalk';
import { LOCAL_DEV_HOST, CERTIFICATE_PATH } from './env';

const { merge } = require('webpack-merge');
const fs = require('fs');
const common = require('./webpack.common.config');

const devtool = 'eval-cheap-module-source-map';

// Check if certificate provided and if not - exit early
if (
  !fs.existsSync(`${CERTIFICATE_PATH}/server.key`) ||
  !fs.existsSync(`${CERTIFICATE_PATH}/server.crt`)
) {
  console.log(
    chalk.red(`ERROR: Can not locate server.key and server.crt in ${CERTIFICATE_PATH} location`),
  );
  console.log(
    chalk.red('Please re-genereate your site certificates by running'),
    'make generate_ssl',
    chalk.red('than re-run the command'),
  );
  process.exit(0);
}

/**
 * Client configuration
 *
 * Client is compiled into multiple chunks that are result to dynamic imports.
 */
export const clientConfig: webpack.Configuration = merge(common.default.clientConfig, {
  mode: 'development',
  devtool,
  devServer: {
    hot: true,
    static: path.join(__dirname, 'dist'),
    compress: true,
    port: 3000,
    host: LOCAL_DEV_HOST,
    historyApiFallback: true,
    server: {
      type: 'https',
      options: {
        key: fs.readFileSync(`${CERTIFICATE_PATH}/server.key`),
        cert: fs.readFileSync(`${CERTIFICATE_PATH}/server.crt`),
      },
    },
    client: {
      logging: 'verbose',
      overlay: { errors: true, warnings: false },
      progress: true,
    },
    devMiddleware: {
      serverSideRender: true,
    },
  },
  optimization: {
    emitOnErrors: false,
    runtimeChunk: 'single',
  },
  plugins: [
    new HTMLWebpackPlugin({
      template: './src/assets/index.html',
      inject: 'body',
      minify: false,
      hash: false,
      showErrors: true,
    }),
  ],
});

/**
 * Server configuration
 *
 * Server bundle is compiled as a CommonJS package that exports an Express middleware
 */
export const serverConfig: webpack.Configuration = merge(common.default.serverConfig, {
  mode: 'development',
  devtool: devtool,
});

export default [clientConfig, serverConfig];
