import { NodeExecutionsById } from 'models/Execution/types';
import { dNode, dTypes } from 'models/Graph/types';
import { Elements, HandleProps } from 'react-flow-renderer';

export interface RFWrapperProps {
    rfGraphJson: Elements;
    backgroundStyle: RFBackgroundProps;
    type: RFGraphTypes;
    onNodeSelectionChanged?: any;
    version?: string;
}

/* Note: extending to allow applying styles directly to handle */
export interface RFHandleProps extends HandleProps {
    style: any;
}

export enum RFGraphTypes {
    main,
    nested,
    static
}

export interface LayoutRCProps {
    setElements: any;
    setLayout: any;
    hasLayout: boolean;
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
    isStaticGraph: boolean;
}

export interface ConvertDagProps {
    root: dNode;
    nodeExecutionsById: any;
    onNodeSelectionChanged: any;
    maxRenderDepth: number;
    isStaticGraph: boolean;
}

export interface DagToFRProps extends ConvertDagProps {
    currentDepth: number;
}

export interface RFCustomData {
    nodeExecutionStatus: NodeExecutionsById;
    text: string;
    handles: [];
    nodeType: dTypes;
    scopedId: string;
    dag: any;
    taskType?: dTypes;
    onNodeSelectionChanged?: any;
}
