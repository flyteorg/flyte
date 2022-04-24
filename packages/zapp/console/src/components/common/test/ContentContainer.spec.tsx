import * as React from 'react';
import * as renderer from 'react-test-renderer';

import { ContentContainer } from '../ContentContainer';

const Content = () => <div>the content</div>;

describe('ContentContainer', () => {
  it('matches the known snapshot', () => {
    expect(
      renderer
        .create(
          <ContentContainer>
            <Content />
          </ContentContainer>,
        )
        .toJSON(),
    ).toMatchSnapshot();
  });

  describe('in centered mode', () => {
    it('matches the known snapshot', () => {
      expect(
        renderer
          .create(
            <ContentContainer center={true}>
              <Content />
            </ContentContainer>,
          )
          .toJSON(),
      ).toMatchSnapshot();
    });
  });
});
