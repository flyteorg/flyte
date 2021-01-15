import { DecoratorFn } from '@storybook/react';
import { ContentContainer } from 'components/common/ContentContainer';
import { NavBar } from 'components/Navigation/NavBar';
import { SideNavigation } from 'components/Navigation/SideNavigation';
import * as React from 'react';

export const withNavigation: DecoratorFn = story => (
    <>
        <NavBar />
        <SideNavigation />
        <ContentContainer sideNav={true}>{story()}</ContentContainer>
    </>
);

export const basicStoryContainer: DecoratorFn = story => (
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
