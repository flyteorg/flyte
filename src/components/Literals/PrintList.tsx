import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { htmlEntities } from './constants';
import { useLiteralStyles } from './styles';

function defaultGetKeyForItem(value: any, index: number) {
  return index;
}

/** Helper component to render a list of items */
export const PrintList: React.FC<{
  getKeyForItem?: (value: any, index: number) => string | number;
  renderValue: (value: any, index: number) => React.ReactNode;
  values: any[];
}> = ({ getKeyForItem = defaultGetKeyForItem, values, renderValue }) => {
  const commonStyles = useCommonStyles();
  const literalStyles = useLiteralStyles();
  const content =
    values.length > 0 ? (
      <ul className={classnames(commonStyles.listUnstyled, literalStyles.nestedContainer)}>
        {values.map((value, idx) => (
          <li key={getKeyForItem(value, idx)}>{renderValue(value, idx)}</li>
        ))}
      </ul>
    ) : null;
  return (
    <>
      <span>{htmlEntities.leftBracket}</span>
      {content}
      <span>{htmlEntities.rightBracket}</span>
    </>
  );
};
