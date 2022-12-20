import classnames from 'classnames';
import { getCacheKey } from 'components/Cache/utils';
import { useCommonStyles } from 'components/common/styles';
import * as scrollbarSize from 'dom-helpers/util/scrollbarSize';
import { NodeExecution } from 'models/Execution/types';
import { dNode } from 'models/Graph/types';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { dateToTimestamp } from 'common/utils';
import * as React from 'react';
import { useMemo, useEffect, useState, useContext } from 'react';
import { isEndNode, isExpanded, isStartNode } from 'components/WorkflowGraph/utils';
import { ExecutionsTableHeader } from './ExecutionsTableHeader';
import { generateColumns } from './nodeExecutionColumns';
import { NoExecutionsContent } from './NoExecutionsContent';
import { useColumnStyles, useExecutionTableStyles } from './styles';
import { NodeExecutionsByIdContext } from '../contexts';
import { convertToPlainNodes } from '../ExecutionDetails/Timeline/helpers';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionRow } from './NodeExecutionRow';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';

interface NodeExecutionsTableProps {
  initialNodes: dNode[];
  filteredNodes?: dNode[];
}

const scrollbarPadding = scrollbarSize();

/**
 * TODO
 * Refactor to avoid code duplication here and in ExecutionTimeline, ie toggleNode, the insides of the effect
 */

/** Renders a table of NodeExecution records. Executions with errors will
 * have an expanadable container rendered as part of the table row.
 * NodeExecutions are expandable and will potentially render a list of child
 * TaskExecutions
 */
export const NodeExecutionsTable: React.FC<NodeExecutionsTableProps> = ({
  initialNodes,
  filteredNodes,
}) => {
  const commonStyles = useCommonStyles();
  const tableStyles = useExecutionTableStyles();
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);
  const { appliedFilters } = useNodeExecutionFiltersState();
  const [originalNodes, setOriginalNodes] = useState<dNode[]>(
    appliedFilters.length > 0 && filteredNodes ? filteredNodes : initialNodes,
  );
  const [showNodes, setShowNodes] = useState<dNode[]>([]);
  const { compiledWorkflowClosure } = useNodeExecutionContext();

  const columnStyles = useColumnStyles();
  // Memoizing columns so they won't be re-generated unless the styles change
  const columns = useMemo(
    () => generateColumns(columnStyles, compiledWorkflowClosure?.primary.template.nodes ?? []),
    [columnStyles],
  );

  useEffect(() => {
    setOriginalNodes(appliedFilters.length > 0 && filteredNodes ? filteredNodes : initialNodes);
    const plainNodes = convertToPlainNodes(originalNodes);
    const updatedShownNodesMap = plainNodes.map((node) => {
      const execution = nodeExecutionsById[node.scopedId];
      return {
        ...node,
        startedAt: execution?.closure.startedAt,
        execution,
      };
    });
    setShowNodes(updatedShownNodesMap);
  }, [initialNodes, filteredNodes, originalNodes, nodeExecutionsById]);

  const toggleNode = (id: string, scopeId: string, level: number) => {
    const searchNode = (nodes: dNode[], nodeLevel: number) => {
      if (!nodes || nodes.length === 0) {
        return;
      }
      for (let i = 0; i < nodes.length; i++) {
        const node = nodes[i];
        if (isStartNode(node) || isEndNode(node)) {
          continue;
        }
        if (node.id === id && node.scopedId === scopeId && nodeLevel === level) {
          nodes[i].expanded = !nodes[i].expanded;
          return;
        }
        if (node.nodes.length > 0 && isExpanded(node)) {
          searchNode(node.nodes, nodeLevel + 1);
        }
      }
    };
    searchNode(originalNodes, 0);
    setOriginalNodes([...originalNodes]);
  };

  return (
    <div className={classnames(tableStyles.tableContainer, commonStyles.flexFill)}>
      <ExecutionsTableHeader columns={columns} scrollbarPadding={scrollbarPadding} />
      <div className={tableStyles.scrollContainer}>
        {showNodes.length > 0 ? (
          showNodes.map((node) => {
            let nodeExecution: NodeExecution;
            if (nodeExecutionsById[node.scopedId]) {
              nodeExecution = nodeExecutionsById[node.scopedId];
            } else {
              nodeExecution = {
                closure: {
                  createdAt: dateToTimestamp(new Date()),
                  outputUri: '',
                  phase: NodeExecutionPhase.UNDEFINED,
                },
                id: {
                  executionId: {
                    domain: node.value?.taskNode?.referenceId?.domain,
                    name: node.value?.taskNode?.referenceId?.name,
                    project: node.value?.taskNode?.referenceId?.project,
                  },
                  nodeId: node.id,
                },
                inputUri: '',
                scopedId: node.scopedId,
              };
            }
            return (
              <NodeExecutionRow
                columns={columns}
                key={getCacheKey(nodeExecution.id)}
                nodeExecution={nodeExecution}
                node={node}
                onToggle={toggleNode}
              />
            );
          })
        ) : (
          <NoExecutionsContent size="large" />
        )}
      </div>
    </div>
  );
};
