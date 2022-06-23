import { env } from 'common/env';

export interface FlyteNavItem {
  title: string;
  url: string;
}

export interface FlyteNavigation {
  color?: string;
  background?: string;
  console?: string;
  items: FlyteNavItem[];
}

export const getFlyteNavigationData = (): FlyteNavigation | undefined => {
  return env.FLYTE_NAVIGATION ? JSON.parse(env.FLYTE_NAVIGATION) : undefined;
};
