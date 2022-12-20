import { getLinkifiedTextChunks, LinkifiedTextChunk } from 'common/linkify';

function text(text: string): LinkifiedTextChunk {
  return { text, type: 'text' };
}

function link(text: string): LinkifiedTextChunk {
  return { text, url: text, type: 'link' };
}

describe('linkify/getLinkifiedTextChunks', () => {
  const testCases: [string, LinkifiedTextChunk[]][] = [
    ['No match expected', [text('No match expected')]],
    ['Points to http://example.com', [text('Points to '), link('http://example.com')]],
    [
      'https://example.com link is at beginning',
      [link('https://example.com'), text(' link is at beginning')],
    ],
    [
      'A link to http://example.com is in the middle',
      [text('A link to '), link('http://example.com'), text(' is in the middle')],
    ],
    [
      'A link at the end to http://example.com',
      [text('A link at the end to '), link('http://example.com')],
    ],
    [
      'A link to http://example.com and another link to https://flyte.org in the middle.',
      [
        text('A link to '),
        link('http://example.com'),
        text(' and another link to '),
        link('https://flyte.org'),
        text(' in the middle.'),
      ],
    ],
  ];
  testCases.forEach(([input, expected]) =>
    it(`correctly splits: ${input}`, () => {
      expect(getLinkifiedTextChunks(input)).toEqual(expected);
    }),
  );
});
