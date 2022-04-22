import * as webpack from 'webpack';
import * as HTMLWebpackPlugin from 'html-webpack-plugin';
import * as path from 'path';
import { processEnv as env, LOCAL_DEV_HOST } from './env';

const { merge } = require('webpack-merge');
const fs = require('fs');
const common = require('./webpack.common.config');

const devtool = 'eval-cheap-module-source-map';

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
        key: fs.readFileSync('script/server.key'),
        cert: fs.readFileSync('script/server.crt'),
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
      clientEnv: common.clientEnv,
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
