import { getAdminApiUrl, getEndpointUrl } from '.';
import { AdminEndpoint, RawEndpoint } from './constants';
import { transformRequestError } from './errors';
import { isObject } from './nodeChecks';

describe('flyte-api/utils', () => {
  it('getEndpointUrl properly uses local or admin domain', () => {
    // when admin domain is not provided - uses localhost
    let result = getEndpointUrl(RawEndpoint.Profile);
    expect(result).toEqual('http://localhost/me');

    // when admin domain is empty - uses localhost
    result = getEndpointUrl(RawEndpoint.Profile, '');
    expect(result).toEqual('http://localhost/me');

    // when admin domain is empty - uses localhost
    result = getEndpointUrl(RawEndpoint.Profile, '');
    expect(result).toEqual('http://localhost/me');

    // when admin domain provided - uses it
    result = getEndpointUrl(RawEndpoint.Profile, 'https://admin.domain.io');
    expect(result).toEqual('https://admin.domain.io/me');
  });

  it('getAdminApiUrl properly adds api/v1 prefix', () => {
    // when admin domain is not provided - uses localhost
    let result = getAdminApiUrl(AdminEndpoint.Version);
    expect(result).toEqual('http://localhost/api/v1/version');

    result = getAdminApiUrl('execution?=filter', 'https://admin.domain.io');
    expect(result).toEqual('https://admin.domain.io/api/v1/execution?=filter');
  });

  it('isObject properly identifies objects', () => {
    // Not an objects
    expect(isObject(null)).toBeFalsy();
    expect(isObject(undefined)).toBeFalsy();
    expect(isObject(3)).toBeFalsy();
    expect(isObject('abc')).toBeFalsy();
    expect(isObject(true)).toBeFalsy();

    // Objects
    expect(isObject({ hi: 'there' })).toBeTruthy();
    expect(isObject([])).toBeTruthy();
    expect(isObject({})).toBeTruthy();
  });

  it('transformRequestError', () => {
    // no status - return item as is
    let result = transformRequestError({ message: 'default' }, '');
    expect(result.message).toEqual('default');

    // 401 - Unauthorised
    result = transformRequestError({ response: { status: 401 }, message: 'default' }, '');
    expect(result.message).toEqual('User is not authorized to view this resource');

    // 404 - Not Found
    result = transformRequestError({ response: { status: 404 }, message: 'default' }, '');
    expect(result.message).toEqual('The requested item could not be found');

    // unnown status - return item as is
    result = transformRequestError({ response: { status: 502 }, message: 'default' }, '');
    expect(result.message).toEqual('default');
  });
});
