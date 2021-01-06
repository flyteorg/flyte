import { Env } from 'config';

/** equivalent to process.env in server and client */
// tslint:disable-next-line:no-any
export const env: Env = Object.assign({}, process.env, window.env);
