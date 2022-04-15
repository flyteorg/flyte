import classnames from 'classnames';
import { sortedObjectEntries } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { Literal, LiteralMap } from 'models/Common/types';
import * as React from 'react';
import { htmlEntities } from './constants';
import { LiteralValue } from './LiteralValue';
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
}> = ({ className, map, showBrackets = false }) => {
  if (!map) {
    return <NoDataIsAvailable />;
  }

  const commonStyles = useCommonStyles();
  const { literals } = map;
  const mapContent = Object.keys(literals).length ? (
    <ul className={classnames(className, commonStyles.textMonospace, commonStyles.listUnstyled)}>
      {sortedObjectEntries(literals).map(([key, value]) => (
        <li key={key}>
          <LiteralValue label={key} literal={value as Literal} />
        </li>
      ))}
    </ul>
  ) : (
    <div className={commonStyles.flexCenter}>
      <NoneTypeValue />
    </div>
  );
  return (
    <>
      {showBrackets && <span>{htmlEntities.leftCurlyBrace}</span>}
      {mapContent}
      {showBrackets && <span>{htmlEntities.rightCurlyBrace}</span>}
    </>
  );
};
