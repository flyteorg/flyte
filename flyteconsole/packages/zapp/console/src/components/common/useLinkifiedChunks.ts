import { getLinkifiedTextChunks } from 'common/linkify';
import { useMemo } from 'react';

export function useLinkifiedChunks(text: string) {
  return useMemo(() => getLinkifiedTextChunks(text), [text]);
}
