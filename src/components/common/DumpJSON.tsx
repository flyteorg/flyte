import { stringifyValue } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';

export const DumpJSON: React.FC<{ value: any }> = ({ value }) => {
  const commonStyles = useCommonStyles();
  return <div className={commonStyles.codeBlock}>{stringifyValue(value)}</div>;
};
