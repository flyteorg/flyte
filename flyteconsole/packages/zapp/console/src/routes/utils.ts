import { env } from 'common/env';

let baseUrl = env.BASE_URL ? env.BASE_URL : '';
if (baseUrl.length && baseUrl[0] !== '/') {
  baseUrl = `/${baseUrl}`;
}

export const makeRoute = (path: string) => `${baseUrl}${path}`;
