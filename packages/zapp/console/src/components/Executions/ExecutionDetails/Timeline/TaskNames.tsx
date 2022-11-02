import * as React from 'react';
import { IconButton, makeStyles, Theme, Tooltip } from '@material-ui/core';

import { RowExpander } from 'components/Executions/Tables/RowExpander';
import { getNodeTemplateName } from 'components/WorkflowGraph/utils';
import { dNode } from 'models/Graph/types';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import { NodeExecutionName } from './NodeExecutionName';
import t from '../strings';

const useStyles = makeStyles((theme: Theme) => ({
  taskNamesList: {
    overflowY: 'scroll',
    flex: 1,
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
    whiteSpace: 'nowrap',
  },
  namesContainerExpander: {
    display: 'flex',
    marginTop: 'auto',
    marginBottom: 'auto',
  },
  namesContainerBody: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-start',
    justifyContent: 'center',
    whiteSpace: 'nowrap',
    height: '100%',
    overflow: 'hidden',
  },
  leaf: {
    width: 30,
  },
}));

interface TaskNamesProps {
  nodes: dNode[];
  onToggle: (id: string, scopeId: string, level: number) => void;
  onAction?: (id: string) => void;
  onScroll?: () => void;
}

export const TaskNames = React.forwardRef<HTMLDivElement, TaskNamesProps>(
  ({ nodes, onScroll, onToggle, onAction }, ref) => {
    const styles = useStyles();

    return (
      <div className={styles.taskNamesList} ref={ref} onScroll={onScroll}>
        {nodes.map((node) => {
          const nodeLevel = node?.level ?? 0;
          return (
            <div
              className={styles.namesContainer}
              key={`level=${nodeLevel}-id=${node.id}-name=${node.scopedId}`}
              data-testid="task-name-item"
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                width: '100%',
                paddingLeft: nodeLevel * 16,
              }}
            >
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'row',
                }}
              >
                <div className={styles.namesContainerExpander}>
                  {node.nodes?.length ? (
                    <RowExpander
                      expanded={node.expanded || false}
                      onClick={() => onToggle(node.id, node.scopedId, nodeLevel)}
                    />
                  ) : (
                    <div className={styles.leaf} />
                  )}
                </div>

                <div className={styles.namesContainerBody}>
                  <NodeExecutionName
                    name={node.name}
                    templateName={getNodeTemplateName(node)}
                    execution={node.execution}
                  />
                </div>
              </div>
              {onAction && (
                <Tooltip title={t('resume')}>
                  <IconButton
                    onClick={() => onAction(node.id)}
                    data-testid={`resume-gate-node-${node.id}`}
                  >
                    <PlayCircleOutlineIcon />
                  </IconButton>
                </Tooltip>
              )}
            </div>
          );
        })}
      </div>
    );
  },
);
