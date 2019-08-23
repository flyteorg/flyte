import { ExpandableMonospaceText } from 'components/common/ExpandableMonospaceText';
import { ExecutionError } from 'models';
import * as React from 'react';
import { useExecutionTableStyles } from './styles';

/** Renders an expandable/collapsible container for an ExecutionErorr, along with
 * a button for copying the error string.
 */
export const ExpandableExecutionError: React.FC<{
    error: ExecutionError;
    onExpandCollapse?(): void;
}> = ({ error, onExpandCollapse }) => {
    const styles = useExecutionTableStyles();
    return (
        <div className={styles.errorContainer}>
            <ExpandableMonospaceText
                onExpandCollapse={onExpandCollapse}
                text={error.message}
            />
        </div>
    );
};
