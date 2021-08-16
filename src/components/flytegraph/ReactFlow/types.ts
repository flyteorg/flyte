import { NodeExecutionsById } from 'models/Execution/types';
import { dTypes } from 'models/Graph/types';
import { Elements, HandleProps } from 'react-flow-renderer';

export interface RFWrapperProps {
    rfGraphJson: Elements;
    backgroundStyle: RFBackgroundProps;
    type: RFGraphTypes;
    onNodeSelectionChanged?: any;
}

/* Note: extending to allow applying styles directly to handle */
export interface RFHandleProps extends HandleProps {
    style: any;
}

export enum RFGraphTypes {
    main,
    nested
}

export interface LayoutRCProps {
    setElements: any;
    setLayout: any;
}

/* React Flow params and styles (background is styles) */
export interface RFBackgroundProps {
    background: any;
    gridColor: string;
    gridSpacing: number;
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
