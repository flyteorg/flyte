import { makeStyles, Theme } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { useCommonStyles } from 'components/common/styles';
import { SelectNodeExecutionLink } from 'components/Executions/Tables/SelectNodeExecutionLink';
import { isEqual } from 'lodash';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { useContext } from 'react';
import { NodeExecutionsTimelineContext } from './context';

interface NodeExecutionTimelineNameData {
  name: string;
  execution: NodeExecution;
}

const useStyles = makeStyles((_theme: Theme) => ({
  selectedExecutionName: {
    fontWeight: 'bold',
  },
}));

export const NodeExecutionName: React.FC<NodeExecutionTimelineNameData> = ({ name, execution }) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const { selectedExecution, setSelectedExecution } = useContext(NodeExecutionsTimelineContext);

  if (!execution) {
    // to avoid crash - disable items which do not have associated execution.
    // as we won't be able to provide task info for them anyway.
    return <Typography variant="body1">{name}</Typography>;
  }

  const isSelected = selectedExecution != null && isEqual(execution.id, selectedExecution);
  return isSelected ? (
    <Typography variant="body1" className={styles.selectedExecutionName}>
      {name}
    </Typography>
  ) : (
    <SelectNodeExecutionLink
      className={commonStyles.primaryLink}
      execution={execution}
      linkText={name}
      setSelectedExecution={setSelectedExecution}
    />
  );
};
