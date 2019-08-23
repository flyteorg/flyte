import * as React from 'react';
import * as renderer from 'react-test-renderer';

import { LoadingSpinner } from '../LoadingSpinner';

describe('LoadingSpinner', () => {
    it('renders large size by default', () => {
        const noProp = renderer.create(<LoadingSpinner />).toJSON();
        const large = renderer.create(<LoadingSpinner size="large" />).toJSON();
        expect(noProp).toEqual(large);
    });

    it('renders small size', () => {
        expect(
            renderer.create(<LoadingSpinner size="small" />).toJSON()
        ).toMatchSnapshot();
    });

    it('renders medium size', () => {
        expect(
            renderer.create(<LoadingSpinner size="medium" />).toJSON()
        ).toMatchSnapshot();
    });

    it('renders large size', () => {
        expect(
            renderer.create(<LoadingSpinner size="large" />).toJSON()
        ).toMatchSnapshot();
    });
});
