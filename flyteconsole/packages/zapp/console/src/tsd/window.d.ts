import { Env } from 'common/env';

declare global {
  export interface Window {
    __INITIAL_DATA__?: {
      config?: Dictionary<object>;
    };
    env: Env;
  }
}
