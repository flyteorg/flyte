import { Env } from 'config';

/** Safely check for server environment */
export function isServer() {
    try {
        return __isServer;
    } catch (e) {
        return false;
    }
}

/** equivalent to process.env in server and client */
// tslint:disable-next-line:no-any
export const env: Env = (isServer() ? process.env : window.env) as any;
