import { makeStyles, Theme } from '@material-ui/core/styles';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import { NonIdealState } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import {
    cannotLaunchWorkflowString,
    requiredInputSuffix,
    unsupportedRequiredInputsString
} from './constants';
import { ParsedInput } from './types';

const useStyles = makeStyles((theme: Theme) => ({
    contentContainer: {
        whiteSpace: 'pre-line',
        textAlign: 'left'
    },
    errorContainer: {
        marginBottom: theme.spacing(2)
    }
}));

function formatLabel(label: string) {
    return label.endsWith(requiredInputSuffix)
        ? label.substring(0, label.length - 1)
        : label;
}

export interface UnsupportedRequiredInputsErrorProps {
    inputs: ParsedInput[];
}
/** An informational error to be shown if a Workflow cannot be launch due to
 * required inputs for which we will not be able to provide a value.
 */
export const UnsupportedRequiredInputsError: React.FC<UnsupportedRequiredInputsErrorProps> = ({
    inputs
}) => {
    const styles = useStyles();
    const commonStyles = useCommonStyles();
    return (
        <NonIdealState
            className={styles.errorContainer}
            icon={ErrorOutline}
            size="medium"
            title={cannotLaunchWorkflowString}
        >
            <div className={styles.contentContainer}>
                <p>{unsupportedRequiredInputsString}</p>
                <ul className={commonStyles.listUnstyled}>
                    {inputs.map(input => (
                        <li
                            key={input.name}
                            className={commonStyles.textMonospace}
                        >
                            {formatLabel(input.label)}
                        </li>
                    ))}
                </ul>
            </div>
        </NonIdealState>
    );
};
