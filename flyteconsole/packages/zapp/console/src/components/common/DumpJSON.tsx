import * as React from 'react';
import { ReactJsonViewWrapper } from 'components/common/ReactJsonView';

export const DumpJSON: React.FC<{ value: any }> = ({ value }) => {
  return <ReactJsonViewWrapper src={value} />;
};
