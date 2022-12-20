import * as LinkifyIt from 'linkify-it';

export const linkify = new LinkifyIt();

export type LinkifiedTextChunkType = 'text' | 'link';
export interface LinkifiedTextChunk {
  type: LinkifiedTextChunkType;
  text: string;
  url?: string;
}

/** Detects any links in the given text and splits the string on the
 * link boundaries, returning an array of text/link entries for each
 * portion of the string.
 */
export function getLinkifiedTextChunks(text: string): LinkifiedTextChunk[] {
  const matches = linkify.match(text);
  if (matches === null) {
    return [{ text, type: 'text' }];
  }

  const chunks: LinkifiedTextChunk[] = [];
  let lastMatchEndIndex = 0;
  matches.forEach((match) => {
    if (lastMatchEndIndex !== match.index) {
      chunks.push({
        text: text.substring(lastMatchEndIndex, match.index),
        type: 'text',
      });
    }
    chunks.push({
      text: match.text,
      type: 'link',
      url: match.url,
    });
    lastMatchEndIndex = match.lastIndex;
  });
  if (lastMatchEndIndex !== text.length) {
    chunks.push({
      text: text.substring(lastMatchEndIndex, text.length),
      type: 'text',
    });
  }

  return chunks;
}
