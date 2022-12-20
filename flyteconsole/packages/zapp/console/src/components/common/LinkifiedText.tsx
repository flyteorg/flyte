import * as React from 'react';
import { NewTargetLink } from './NewTargetLink';
import { useLinkifiedChunks } from './useLinkifiedChunks';

/** Renders the given text with any inline valid URLs rendered as
 * `NewTargetLink`s.
 */
export const LinkifiedText: React.FC<{ text?: string }> = ({ text = '' }) => {
  const chunks = useLinkifiedChunks(text);
  return (
    <>
      {chunks.map((chunk, idx) => {
        const key = `${chunk.type}-${idx}`;
        if (chunk.type === 'link') {
          return (
            <NewTargetLink key={key} inline={true} href={chunk.url}>
              {chunk.text}
            </NewTargetLink>
          );
        }
        return <span key={key}>{chunk.text}</span>;
      })}
    </>
  );
};
