import { getEndpointUrl } from '../utils';
import { RawEndpoint } from '../utils/constants';

export interface LoginStatus {
  expired: boolean;
  setExpired(expired: boolean): void;
}

export const defaultLoginStatus: LoginStatus = {
  expired: true,
  setExpired: () => {
    /** Do nothing */
  },
};

/** Constructs a url for redirecting to the Admin login endpoint and returning
 * to the current location after completing the flow.
 */
export function getLoginUrl(adminUrl?: string, redirectUrl: string = window.location.href) {
  const baseUrl = getEndpointUrl(RawEndpoint.Login, adminUrl);
  return `${baseUrl}?redirect_url=${redirectUrl}`;
}
