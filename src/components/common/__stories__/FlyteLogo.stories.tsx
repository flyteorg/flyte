import { makeStyles, Theme } from '@material-ui/core/styles';
import { storiesOf } from '@storybook/react';
import { basicStoryContainer } from '__stories__/decorators';
import * as React from 'react';
import { FlyteLogo } from '../FlyteLogo';

const useStyles = makeStyles((theme: Theme) => ({
    darkContainer: {
        backgroundColor: theme.palette.secondary.main,
        height: '100%',
        padding: theme.spacing(2),
        width: '100%'
    },
    lightContainer: {
        height: '100%',
        padding: theme.spacing(2),
        width: '100%'
    }
}));

const stories = storiesOf('Common/FlyteLogo', module);
stories.addDecorator(basicStoryContainer);

stories.add('Dark', () => (
    <div className={useStyles().darkContainer}>
        <FlyteLogo variant="dark" size={64} />
    </div>
));

stories.add('Light', () => (
    <div className={useStyles().lightContainer}>
        <FlyteLogo variant="light" size={64} />
    </div>
));
