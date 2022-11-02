import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { workflowNoInputsString } from './constants';
import t from './strings';

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    marginBottom: theme.spacing(1),
    marginTop: theme.spacing(1),
  },
}));

export interface NoInputsProps {
  variant: 'workflow' | 'task';
}
/** An informational message to be shown if a Workflow or Task does not need any
 * input values.
 */
export const NoInputsNeeded: React.FC<NoInputsProps> = ({ variant }) => {
  const commonStyles = useCommonStyles();
  return (
    <Typography
      align="center"
      className={classnames(commonStyles.hintText, useStyles().root)}
      variant="body2"
    >
      {variant === 'workflow' ? workflowNoInputsString : t('taskNoInputsString')}
    </Typography>
  );
};
