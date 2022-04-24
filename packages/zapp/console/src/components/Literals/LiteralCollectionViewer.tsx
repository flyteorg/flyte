import { Literal, LiteralCollection } from 'models/Common/types';
import * as React from 'react';
import { LiteralValue } from './LiteralValue';
import { PrintList } from './PrintList';

function renderCollectionItem(value: any, index: number) {
  return <LiteralValue label={index} literal={value as Literal} />;
}

/** Renders a LiteralCollection as a formatted list */
export const LiteralCollectionViewer: React.FC<{
  collection: LiteralCollection;
}> = ({ collection }) => (
  <PrintList values={collection.literals} renderValue={renderCollectionItem} />
);
