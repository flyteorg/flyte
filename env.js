/**
 * @file Provides environment variables
 * Do not import this file into client/server code. Use common/env.ts instead
 */

/** Current environment environment. "development", "test" or "production" */
const NODE_ENV = process.env.NODE_ENV || 'development';

// If this is unset, API calls will default to the same host used to serve this app
const ADMIN_API_URL = process.env.ADMIN_API_URL;
// Use this to create SSL server
const ADMIN_API_USE_SSL = process.env.ADMIN_API_USE_SSL || 'http';

const BASE_URL = process.env.BASE_URL || '';
const CORS_PROXY_PREFIX = process.env.CORS_PROXY_PREFIX || '/cors_proxy';

// Defines a file to be required which will provide implementations for
// any user-definable code.
const PLUGINS_MODULE = process.env.PLUGINS_MODULE;

// Optional URL which will provide information about system status. This is used
// to show informational banners to the user.
// If it has no protocol, it will be treated as relative to window.location.origin
const STATUS_URL = process.env.STATUS_URL;

// Configure Google Analytics
const ENABLE_GA = process.env.ENABLE_GA || false;
const GA_TRACKING_ID = process.env.GA_TRACKING_ID || 'G-0QW4DJWJ20';

module.exports = {
    ADMIN_API_URL,
    ADMIN_API_USE_SSL,
    BASE_URL,
    CORS_PROXY_PREFIX,
    NODE_ENV,
    PLUGINS_MODULE,
    STATUS_URL,
    ENABLE_GA,
    GA_TRACKING_ID,
    processEnv: {
        ADMIN_API_URL,
        BASE_URL,
        CORS_PROXY_PREFIX,
        ENABLE_GA,
        GA_TRACKING_ID,
        NODE_ENV,
        STATUS_URL
    }
};
