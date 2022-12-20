import * as React from 'react';
import { RouteComponentProps } from 'react-router-dom';

// Wraps a component which expects route params specified by the interface
// ParamsType and returns a HoC which will extract the params from props.match.params
// and render the passed component with those params as top-level props.
export function withRouteParams<ParamsType>(
  Component: React.ComponentType<ParamsType>,
): React.FunctionComponent<RouteComponentProps<ParamsType>> {
  return ({ match }) => {
    return <Component {...match.params} />;
  };
}
