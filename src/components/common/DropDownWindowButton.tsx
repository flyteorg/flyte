import Button, { ButtonProps } from '@material-ui/core/Button';
import Popover from '@material-ui/core/Popover';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';

const { useRef, useState } = React;

type ButtonVariant = 'primary' | 'dangerous';
export type FormCloseHandler = (e: object, reason: string) => void;

interface DropDownWindowButtonProps {
  className?: string;
  text: string;
  variant?: ButtonVariant;
  renderContent: (onClose: FormCloseHandler) => JSX.Element;
}

/** A button which manages the state and popover interaction for showing a small
 * drop-down window.
 */
export const DropDownWindowButton: React.FC<DropDownWindowButtonProps> = ({
  className,
  renderContent,
  text,
  variant = 'primary',
}) => {
  const commonStyles = useCommonStyles();
  const [showingWindow, setShowingWindow] = useState(false);
  const buttonRef = useRef(null);
  let buttonProps: ButtonProps;
  if (variant === 'dangerous') {
    buttonProps = {
      className: classnames(commonStyles.buttonWhiteOutlined, commonStyles.dangerousButton),
      color: 'default',
    };
  } else {
    buttonProps = {
      className: commonStyles.buttonWhiteOutlined,
      color: 'primary',
    };
  }

  const showWindow = () => setShowingWindow(true);
  const closeForm: FormCloseHandler = (e: object, reason: string) => {
    if (reason !== 'backdropClick') {
      setShowingWindow(false);
    }
  };

  return (
    <div className={className}>
      <Button {...buttonProps} onClick={showWindow} ref={buttonRef} size="small" variant="outlined">
        {text}
      </Button>
      <Popover
        anchorEl={buttonRef.current}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        elevation={4}
        onClose={closeForm as React.ReactEventHandler<{}>}
        open={showingWindow}
        PaperProps={{ className: commonStyles.smallDropdownWindow }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        {showingWindow ? renderContent(closeForm) : null}
      </Popover>
    </div>
  );
};
