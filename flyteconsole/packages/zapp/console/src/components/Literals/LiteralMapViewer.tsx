import { ReactJsonViewWrapper } from 'components/common/ReactJsonView';
import { LiteralMap } from 'models/Common/types';
import * as React from 'react';
import { transformLiterals } from './helpers';
import { NoneTypeValue } from './Scalar/NoneTypeValue';

export const NoDataIsAvailable = () => {
  return (
    <p>
      <em>No data is available.</em>
    </p>
  );
};

/** Renders a LiteralMap as a formatted object */
export const LiteralMapViewer: React.FC<{
  className?: string;
  map: LiteralMap | null;
  showBrackets?: boolean;
}> = ({ map }) => {
  if (!map) {
    return <NoDataIsAvailable />;
  }

  const { literals } = map;

  if (!Object.keys(literals).length) {
    return <NoneTypeValue />;
  }

  const transformedLiterals = transformLiterals(literals);

  return <ReactJsonViewWrapper src={transformedLiterals} />;
};
