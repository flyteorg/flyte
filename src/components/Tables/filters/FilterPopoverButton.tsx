import { Button, Popover } from '@material-ui/core';
import { fade, makeStyles, Theme } from '@material-ui/core/styles';
import Close from '@material-ui/icons/Close';
import ExpandMore from '@material-ui/icons/ExpandMore';
import classnames from 'classnames';
import {
    buttonHoverColor,
    interactiveTextBackgroundColor,
    interactiveTextColor
} from 'components/Theme/constants';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => {
    const buttonInteractiveStyles = {
        borderColor: theme.palette.text.primary,
        color: theme.palette.text.primary
    };
    const horizontalButtonPadding = theme.spacing(1.5);
    return {
        button: {
            backgroundColor: theme.palette.common.white,
            borderColor: theme.palette.divider,
            color: theme.palette.text.secondary,
            height: theme.spacing(4),
            fontWeight: 'bold',
            lineHeight: '1.1875rem',
            padding: `${theme.spacing(0.375)}px ${horizontalButtonPadding}px`,
            textTransform: 'none',
            '&.active, &.active:hover, &.active.open': {
                backgroundColor: interactiveTextBackgroundColor,
                borderColor: 'transparent',
                color: interactiveTextColor
            },
            '&:hover': {
                ...buttonInteractiveStyles,
                backgroundColor: buttonHoverColor
            },
            '&.open': buttonInteractiveStyles
        },
        buttonIcon: {
            marginLeft: theme.spacing(1),
            marginRight: -horizontalButtonPadding / 2
        },
        resetIcon: {
            cursor: 'pointer',
            '&:hover': {
                color: fade(interactiveTextColor, 0.4)
            }
        },
        popoverContent: {
            border: `1px solid ${theme.palette.divider}`,
            borderRadius: 4,
            marginTop: theme.spacing(0.25),
            padding: `${theme.spacing(2)}px ${theme.spacing(1.5)}px`
        }
    };
});

export interface FilterPopoverButtonProps {
    active?: boolean;
    buttonText: string;
    className?: string;
    onClick?: React.MouseEventHandler<HTMLButtonElement>;
    onReset?: () => void;
    open: boolean;
    renderContent: () => JSX.Element;
}

/** Renders a common filter button with shared behavior for active/hover states,
 * a reset icon, and rendering the provided content in a `Popover`. The state
 * for this button can be mostly generated using the `useFilterButtonState` hook,
 * but will generally be included as part of a bigger filter state such as
 * `SingleSelectFilterState`.
 */
export const FilterPopoverButton: React.FC<FilterPopoverButtonProps> = ({
    active,
    buttonText,
    className,
    onClick,
    onReset,
    open,
    renderContent
}) => {
    const buttonRef = React.useRef<HTMLButtonElement>(null);
    const styles = useStyles();

    let iconContent: JSX.Element;

    if (active && onReset) {
        const onClickReset = (event: React.MouseEvent) => {
            event.stopPropagation();
            onReset();
        };
        iconContent = (
            <Close
                className={classnames(styles.buttonIcon, styles.resetIcon)}
                fontSize="inherit"
                onClick={onClickReset}
            />
        );
    } else {
        iconContent = (
            <ExpandMore className={styles.buttonIcon} fontSize="inherit" />
        );
    }

    return (
        <div className={className}>
            <Button
                disableRipple={true}
                disableTouchRipple={true}
                className={classnames(styles.button, { active, open })}
                ref={buttonRef}
                onClick={onClick}
                variant="outlined"
            >
                {buttonText}
                {iconContent}
            </Button>
            <Popover
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left'
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'left'
                }}
                anchorEl={buttonRef.current}
                elevation={1}
                onClose={onClick}
                open={open}
                PaperProps={{ className: styles.popoverContent }}
            >
                {renderContent()}
            </Popover>
        </div>
    );
};
