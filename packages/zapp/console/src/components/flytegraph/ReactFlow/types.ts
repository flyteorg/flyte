import { CatalogCacheStatus } from 'models/Execution/enums';
import { NodeExecutionsById } from 'models/Execution/types';
import { dNode, dTypes } from 'models/Graph/types';
import { HandleProps } from 'react-flow-renderer';

export interface RFWrapperProps {
  rfGraphJson: any;
  backgroundStyle: RFBackgroundProps;
  type?: RFGraphTypes;
  currentNestedView?: any;
  onNodeSelectionChanged?: any;
  nodeExecutionsById?: any;
  version?: string;
}

/* Note: extending to allow applying styles directly to handle */
export interface RFHandleProps extends HandleProps {
  style: any;
}

export enum RFGraphTypes {
  main,
  nested,
  static,
}

export interface LayoutRCProps {
  setPositionedElements: any;
  graphData: any;
  nodeExecutionsById: any;
}

/* React Flow params and styles (background is styles) */
export interface RFBackgroundProps {
  background: any;
  gridColor: string;
  gridSpacing: number;
}

export interface BuildRFNodeProps {
  dNode: dNode;
  dag?: any[];
  nodeExecutionsById: any;
  typeOverride: dTypes | null;
  onNodeSelectionChanged: any;
  onAddNestedView?: any;
  onRemoveNestedView?: any;
  currentNestedView?: any;
  isStaticGraph: boolean;
}

export interface ConvertDagProps {
  root: dNode;
  nodeExecutionsById: any;
  onNodeSelectionChanged: any;
  onRemoveNestedView?: any;
  onAddNestedView?: any;
  currentNestedView?: any;
  maxRenderDepth: number;
  isStaticGraph?: boolean;
}

export interface DagToReactFlowProps extends ConvertDagProps {
  currentDepth: number;
  parents: any;
}

export interface RFCustomData {
  nodeExecutionStatus: NodeExecutionsById;
  text: string;
  handles: [];
  nodeType: dTypes;
  scopedId: string;
  dag: any;
  taskType?: dTypes;
  cacheStatus?: CatalogCacheStatus;
  onNodeSelectionChanged?: any;
  onAddNestedView: any;
  onRemoveNestedView: any;
}
