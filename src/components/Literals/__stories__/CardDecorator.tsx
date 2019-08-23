import { Card, CardContent } from '@material-ui/core';
import { StoryDecorator } from '@storybook/react';
import * as React from 'react';

import { useCommonStyles } from 'components/common/styles';

/** Shared decorator for Literal stories which places each story inside a Card
 * and sets a monospace font for better readability.
 */
export const CardDecorator: StoryDecorator = story => (
    <Card>
        <CardContent>
            <div className={useCommonStyles().textMonospace}>{story()}</div>
        </CardContent>
    </Card>
);
