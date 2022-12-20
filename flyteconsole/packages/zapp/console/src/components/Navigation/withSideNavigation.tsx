import { ContentContainer, ContentContainerProps } from 'components/common/ContentContainer';
import * as React from 'react';
import { SideNavigation } from './SideNavigation';

export function withSideNavigation<P>(
  WrappedComponent: React.ComponentType<P>,
  contentContainerProps: ContentContainerProps = {},
) {
  return (props: P) => (
    <>
      <SideNavigation />
      <ContentContainer {...contentContainerProps} sideNav={true}>
        <WrappedComponent {...props} />
      </ContentContainer>
    </>
  );
}
