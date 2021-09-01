import { Admin } from 'flyteidl';
import { ExpandableMonospaceText } from 'components/common/ExpandableMonospaceText';
import { ExecutionError } from 'models/Execution/types';
import * as React from 'react';
import { useExecutionTableStyles } from './styles';

/** Renders an expandable/collapsible container for an ExecutionErorr, along with
 * a button for copying the error string.
 */
export const ExpandableExecutionError: React.FC<{
    abortMetadata?: Admin.IAbortMetadata;
    error?: ExecutionError;
    initialExpansionState?: boolean;
    onExpandCollapse?(expanded: boolean): void;
}> = ({
    abortMetadata,
    error,
    initialExpansionState = false,
    onExpandCollapse
}) => {
    const styles = useExecutionTableStyles();
    return (
        <div className={styles.errorContainer}>
            <ExpandableMonospaceText
                onExpandCollapse={onExpandCollapse}
                initialExpansionState={initialExpansionState}
                text={abortMetadata?.cause ?? error?.message ?? ''}
            />
        </div>
    );
};
