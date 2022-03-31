/* eslint import/no-dynamic-require: 'off', no-console: 'off' */
const env = require('./env');

const { middleware } = env.PLUGINS_MODULE ? require(env.PLUGINS_MODULE) : {};

if (Array.isArray(middleware)) {
  module.exports.applyMiddleware = (app) => middleware.forEach((m) => m(app));
} else if (middleware !== undefined) {
  console.log(`Expected middleware to be of type Array, but got ${middleware} instead`);
}
