import * as React from 'react';
import { makeStyles, Theme, Typography } from '@material-ui/core';

import { RowExpander } from 'components/Executions/Tables/RowExpander';
import { NodeExecutionName } from './NodeExecutionName';
import { NodeExecutionsTimelineContext } from './context';
import { getNodeTemplateName } from 'components/WorkflowGraph/utils';
import { dNode } from 'models/Graph/types';

const useStyles = makeStyles((theme: Theme) => ({
  taskNamesList: {
    overflowY: 'scroll',
    flex: 1
  },
  namesContainer: {
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'left',
    padding: '0 10px',
    height: 56,
    width: 256,
    borderBottom: `1px solid ${theme.palette.divider}`,
    whiteSpace: 'nowrap'
  },
  namesContainerExpander: {
    display: 'flex',
    marginTop: 'auto',
    marginBottom: 'auto'
  },
  namesContainerBody: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
    justifyContent: 'center',
    whiteSpace: 'nowrap',
    height: '100%',
    overflow: 'hidden'
  },
  displayName: {
    marginTop: 4,
    textOverflow: 'ellipsis',
    width: '100%',
    overflow: 'hidden'
  },
  leaf: {
    width: 30
  }
}));

interface TaskNamesProps {
  nodes: dNode[];
  onToggle: (id: string, scopeId: string) => void;
}

export const TaskNames = React.forwardRef<HTMLDivElement, TaskNamesProps>((props, ref) => {
  const state = React.useContext(NodeExecutionsTimelineContext);
  const { nodes, onToggle } = props;
  const styles = useStyles();

  return (
    <div className={styles.taskNamesList} ref={ref}>
      {nodes.map(node => {
        const templateName = getNodeTemplateName(node);
        return (
          <div
            className={styles.namesContainer}
            key={`task-name-${node.scopedId}`}
            style={{ paddingLeft: (node?.level || 0) * 16 }}
          >
            <div className={styles.namesContainerExpander}>
              {node.nodes?.length ? (
                <RowExpander expanded={node.expanded || false} onClick={() => onToggle(node.id, node.scopedId)} />
              ) : (
                <div className={styles.leaf} />
              )}
            </div>

            <div className={styles.namesContainerBody}>
              <NodeExecutionName
                name={node.name}
                execution={node.execution!} //node.execution currently accessible only for root nodes
                state={state}
              />
              <Typography variant="subtitle1" color="textSecondary" className={styles.displayName}>
                {templateName}
              </Typography>
            </div>
          </div>
        );
      })}
    </div>
  );
});
