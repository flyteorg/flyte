import { DropDownWindowButton, FormCloseHandler } from 'components/common/DropDownWindowButton';
import * as React from 'react';
import { TerminateExecutionForm } from './TerminateExecutionForm';

interface TerminateExecutionButtonProps {
  className?: string;
}

/** A button which manages the state and popover interaction used by a
 * `TerminateExecutionForm`
 */
export const TerminateExecutionButton: React.FC<TerminateExecutionButtonProps> = ({
  className,
}) => {
  const renderContent = (onClose: FormCloseHandler) => <TerminateExecutionForm onClose={onClose} />;
  return (
    <DropDownWindowButton
      className={className}
      renderContent={renderContent}
      text="Terminate"
      variant="dangerous"
    />
  );
};
