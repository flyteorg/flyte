import { StoryDecorator } from '@storybook/react';
import * as React from 'react';

import { ContentContainer } from 'components/common';
import { NavBar } from 'components/Navigation/NavBar';
import { SideNavigation } from 'components/Navigation/SideNavigation';

export const withNavigation: StoryDecorator = story => (
    <>
        <NavBar />
        <SideNavigation />
        <ContentContainer sideNav={true}>{story()}</ContentContainer>
    </>
);

export const basicStoryContainer: StoryDecorator = story => (
    <div
        style={{
            display: 'flex',
            height: '100vh',
            padding: 20,
            width: '100vw'
        }}
    >
        {story()}
    </div>
);
