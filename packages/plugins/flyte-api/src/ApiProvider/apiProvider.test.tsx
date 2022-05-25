import * as React from 'react';
import { render } from '@testing-library/react';
import { FlyteApiProvider, useFlyteApi } from '.';
import { AdminEndpoint } from '../utils/constants';
import { getLoginUrl } from './login';

const MockCoponent = () => {
  const context = useFlyteApi();

  return (
    <>
      <div>{context.getProfileUrl()}</div>
      <div>{context.getAdminApiUrl('/magic')}</div>
      <div>{context.getLoginUrl()}</div>
    </>
  );
};

describe('fltyte-api/ApiProvider', () => {
  it('getLoginUrl properly adds redirect url', () => {
    const result = getLoginUrl(AdminEndpoint.Version, `http://some.nonsense`);
    expect(result).toEqual('/version/login?redirect_url=http://some.nonsense');
  });

  it('If FlyteApiContext is not defined, returned URL uses default value', () => {
    const { getAllByText } = render(<MockCoponent />);
    expect(getAllByText('#').length).toBe(3);
  });

  it('If FlyteApiContext is defined, but flyteApiDomain is not point to localhost', () => {
    const { getByText } = render(
      <FlyteApiProvider>
        <MockCoponent />
      </FlyteApiProvider>,
    );
    expect(getByText('http://localhost/me')).toBeInTheDocument();
  });

  it('If FlyteApiContext provides flyteApiDomain value', () => {
    const { getByText } = render(
      <FlyteApiProvider flyteApiDomain="https://some.domain.here">
        <MockCoponent />
      </FlyteApiProvider>,
    );
    expect(getByText('https://some.domain.here/me')).toBeInTheDocument();
  });
});
