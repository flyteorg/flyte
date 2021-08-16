import { CompiledNode } from 'models/Node/types';
import {
    WorkflowTemplate,
    CompiledWorkflow,
    CompiledWorkflowClosure
} from 'models/Workflow/types';

export const workflowData = require('models/__mocks__/simpleWorkflowClosure.json');

export const mockCompiledWorkflowClosure: CompiledWorkflowClosure =
    workflowData.compiledWorkflow;

export const mockCompiledWorkflow: CompiledWorkflow =
    mockCompiledWorkflowClosure.primary;

export const mockTemplate: WorkflowTemplate =
    mockCompiledWorkflowClosure.primary.template;

export const mockNodesList: CompiledNode[] = mockTemplate.nodes;
export const mockCompiledStartNode: CompiledNode = mockNodesList[0];
export const mockCompiledEndNode: CompiledNode = mockNodesList[1];
export const mockCompiledTaskNode: CompiledNode = mockNodesList[2];

// const subWorkflow: CompiledWorkflow[] = [{
//     template:{
//         id:{

//         }
//     },
//     connections: {}
// }]
// export const mockSubworkflow: CompiledWorkflow =
