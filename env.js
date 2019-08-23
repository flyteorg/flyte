/**
 * @file Provides environment variables
 * Do not import this file into client/server code. Use common/env.ts instead
 */

/** Current environment environment. "development", "test" or "production" */
const NODE_ENV = process.env.NODE_ENV || 'development';

// If this is unset, API calls will default to the same host used to serve this app
const ADMIN_API_URL = process.env.ADMIN_API_URL;

const BASE_URL = process.env.BASE_URL || '';
const CORS_PROXY_PREFIX = process.env.CORS_PROXY_PREFIX || '/cors_proxy';

const CONFIG_DIR = process.env.CONFIG_DIR || './config';
const CONFIG_CACHE_TTL_SECONDS = process.env.CONFIG_CACHE_TTL_SECONDS || '60';

module.exports = {
    ADMIN_API_URL,
    BASE_URL,
    // Config env isn't sent to the client, so excluded from processEnv
    CONFIG_CACHE_TTL_SECONDS,
    CONFIG_DIR,
    CORS_PROXY_PREFIX,
    NODE_ENV,
    processEnv: {
        ADMIN_API_URL,
        BASE_URL,
        CORS_PROXY_PREFIX,
        NODE_ENV
    }
};
