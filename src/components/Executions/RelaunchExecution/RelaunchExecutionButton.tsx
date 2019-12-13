import {
    DropDownWindowButton,
    FormCloseHandler
} from 'components/common/DropDownWindowButton';
import { Execution } from 'models';
import * as React from 'react';
import { RelaunchExecutionForm } from './RelaunchExecutionForm';

interface RelaunchExecutionButtonProps {
    className?: string;
    execution: Execution;
}

/** A button which manages the state and popover interaction used by a
 * `RelaunchExecutionForm`
 */
export const RelaunchExecutionButton: React.FC<RelaunchExecutionButtonProps> = ({
    className,
    execution
}) => {
    const renderContent = (onClose: FormCloseHandler) => (
        <RelaunchExecutionForm execution={execution} onClose={onClose} />
    );
    return (
        <DropDownWindowButton
            className={className}
            renderContent={renderContent}
            text="Relaunch"
            variant="primary"
        />
    );
};
