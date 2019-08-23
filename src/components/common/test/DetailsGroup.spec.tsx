import * as React from 'react';
import * as renderer from 'react-test-renderer';

import { DetailsGroup } from '../DetailsGroup';

describe('DetailsGroup', () => {
    it('matches the known snapshot', () => {
        const items = [
            { name: 'item1', content: 'a simple string' },
            { name: 'item2', content: <div>A React node</div> }
        ];
        expect(
            renderer.create(<DetailsGroup items={items} />).toJSON()
        ).toMatchSnapshot();
    });
});
