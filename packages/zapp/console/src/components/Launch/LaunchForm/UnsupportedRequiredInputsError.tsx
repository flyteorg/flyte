import { makeStyles, Theme } from '@material-ui/core/styles';
import ErrorOutline from '@material-ui/icons/ErrorOutline';
import { NonIdealState } from 'components/common/NonIdealState';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import {
  cannotLaunchTaskString,
  cannotLaunchWorkflowString,
  requiredInputSuffix,
  taskUnsupportedRequiredInputsString,
  workflowUnsupportedRequiredInputsString,
} from './constants';
import { ParsedInput } from './types';

const useStyles = makeStyles((theme: Theme) => ({
  contentContainer: {
    whiteSpace: 'pre-line',
    textAlign: 'left',
  },
  errorContainer: {
    marginBottom: theme.spacing(2),
  },
}));

function formatLabel(label: string) {
  return label.endsWith(requiredInputSuffix) ? label.substring(0, label.length - 1) : label;
}

export interface UnsupportedRequiredInputsErrorProps {
  inputs: ParsedInput[];
  variant: 'workflow' | 'task';
}
/** An informational error to be shown if a Workflow cannot be launch due to
 * required inputs for which we will not be able to provide a value.
 */
export const UnsupportedRequiredInputsError: React.FC<UnsupportedRequiredInputsErrorProps> = ({
  inputs,
  variant,
}) => {
  const styles = useStyles();
  const commonStyles = useCommonStyles();
  const [titleString, errorString] =
    variant === 'workflow'
      ? [cannotLaunchWorkflowString, workflowUnsupportedRequiredInputsString]
      : [cannotLaunchTaskString, taskUnsupportedRequiredInputsString];
  return (
    <NonIdealState
      className={styles.errorContainer}
      icon={ErrorOutline}
      size="medium"
      title={titleString}
    >
      <div className={styles.contentContainer}>
        <p>{errorString}</p>
        <ul className={commonStyles.listUnstyled}>
          {inputs.map((input) => (
            <li key={input.name} className={commonStyles.textMonospace}>
              {formatLabel(input.label)}
            </li>
          ))}
        </ul>
      </div>
    </NonIdealState>
  );
};
