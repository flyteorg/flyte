import { makeStyles, Theme } from '@material-ui/core/styles';
import * as classnames from 'classnames';
import { contentContainerId } from 'common/constants';
import {
    contentMarginGridUnits,
    maxContainerGridWidth,
    navbarGridHeight,
    sideNavGridWidth
} from 'common/layout';
import * as React from 'react';
import { detailsPanelWidth } from './constants';
import { DetailsPanel } from './DetailsPanel';
import { ErrorBoundary } from './ErrorBoundary';

enum ContainerClasses {
    Centered = 'centered',
    NoMargin = 'nomargin',
    WithDetailsPanel = 'withDetailsPanel',
    WithSideNav = 'withSideNav'
}

const useStyles = makeStyles((theme: Theme) => {
    const contentMargin = `${theme.spacing(contentMarginGridUnits)}px`;
    const spacerHeight = `${theme.spacing(navbarGridHeight)}px`;
    return {
        root: {
            display: 'flex',
            flexDirection: 'column',
            minHeight: '100vh',
            padding: `${spacerHeight} ${contentMargin} 0 ${contentMargin}`,
            [`&.${ContainerClasses.NoMargin}`]: {
                margin: 0,
                padding: `${spacerHeight} 0 0 0`
            },
            [`&.${ContainerClasses.Centered}`]: {
                margin: '0 auto',
                maxWidth: theme.spacing(maxContainerGridWidth)
            },
            [`&.${ContainerClasses.WithDetailsPanel}`]: {
                paddingRight: 0
            },
            [`&.${ContainerClasses.WithSideNav}`]: {
                marginLeft: theme.spacing(sideNavGridWidth)
            }
        }
    };
});

export interface ContentContainerProps
    extends React.AllHTMLAttributes<HTMLDivElement> {
    /** Renders content centered in the page. Usually should not be combined
     * with other modifiers
     */
    center?: boolean;
    /** Controls rendering of a `DetailsPanel` instance, which child content
     * can render into using `DetailsPanelContent`
     */
    detailsPanel?: boolean;
    /** Renders the container with no margin or padding.  Usually should not be combined
     * with other modifiers */
    noMargin?: boolean;
    /** Whether to include spacing for the SideNavigation component */
    sideNav?: boolean;
}

/** Defines the main content container for the application. Only one of these
 * should be present at a time, as it uses an id to assist with some
 * react-virtualized behavior.
 */
export const ContentContainer: React.FC<ContentContainerProps> = props => {
    const styles = useStyles();
    const {
        center = false,
        noMargin = false,
        className: additionalClassName,
        children,
        detailsPanel = false,
        sideNav = false,
        ...restProps
    } = props;

    const className = classnames(styles.root, additionalClassName, {
        [ContainerClasses.Centered]: center,
        [ContainerClasses.NoMargin]: noMargin,
        [ContainerClasses.WithDetailsPanel]: detailsPanel,
        [ContainerClasses.WithSideNav]: sideNav
    });

    const marginRight = detailsPanel ? detailsPanelWidth : 'auto';

    return (
        <>
            <div
                {...restProps}
                className={className}
                id={contentContainerId}
                style={{ marginRight }}
            >
                <ErrorBoundary>{children}</ErrorBoundary>
            </div>
            {detailsPanel && <DetailsPanel />}
        </>
    );
};
