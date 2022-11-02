import { makeStyles, Theme } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { useCommonStyles } from 'components/common/styles';
import { useNodeExecutionContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { SelectNodeExecutionLink } from 'components/Executions/Tables/SelectNodeExecutionLink';
import { isEqual } from 'lodash';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { useContext, useEffect, useState } from 'react';
import { DetailsPanelContext } from '../DetailsPanelContext';

interface NodeExecutionTimelineNameData {
  name: string;
  templateName?: string;
  execution?: NodeExecution;
}

const useStyles = makeStyles((_theme: Theme) => ({
  selectedExecutionName: {
    fontWeight: 'bold',
  },
  displayName: {
    marginTop: 4,
    textOverflow: 'ellipsis',
    width: '100%',
    overflow: 'hidden',
  },
}));

export const NodeExecutionName: React.FC<NodeExecutionTimelineNameData> = ({
  name,
  templateName,
  execution,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const { getNodeExecutionDetails } = useNodeExecutionContext();
  const { selectedExecution, setSelectedExecution } = useContext(DetailsPanelContext);
  const [displayName, setDisplayName] = useState<string | undefined>();

  useEffect(() => {
    let isCurrent = true;
    getNodeExecutionDetails(execution).then((res) => {
      if (isCurrent) {
        setDisplayName(res.displayName);
      }
    });
    return () => {
      isCurrent = false;
    };
  });

  if (!execution) {
    // to avoid crash - disable items which do not have associated execution.
    // as we won't be able to provide task info for them anyway.
    return <Typography variant="body1">{name}</Typography>;
  }
  const isSelected = selectedExecution != null && isEqual(execution.id, selectedExecution);

  const defaultName = displayName ?? name;
  const truncatedName = defaultName?.split('.').pop() || defaultName;

  return (
    <>
      {isSelected || execution.closure.phase === NodeExecutionPhase.UNDEFINED ? (
        <Typography variant="body1" className={styles.selectedExecutionName}>
          {truncatedName}
        </Typography>
      ) : (
        <SelectNodeExecutionLink
          className={commonStyles.primaryLink}
          execution={execution}
          linkText={truncatedName || ''}
          setSelectedExecution={setSelectedExecution}
        />
      )}
      {templateName && (
        <Typography variant="subtitle1" color="textSecondary" className={styles.displayName}>
          {templateName}
        </Typography>
      )}
    </>
  );
};
