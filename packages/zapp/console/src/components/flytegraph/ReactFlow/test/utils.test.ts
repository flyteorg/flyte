import { nodeExecutionPhaseConstants } from 'components/Executions/constants';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { COLOR_NOT_EXECUTED, getStatusColor } from '../utils';

describe('getStatusColor', () => {
  describe.each`
    nodeExecutionStatus             | expected
    ${undefined}                    | ${COLOR_NOT_EXECUTED}
    ${NodeExecutionPhase.FAILED}    | ${nodeExecutionPhaseConstants[NodeExecutionPhase.FAILED].nodeColor}
    ${NodeExecutionPhase.FAILING}   | ${nodeExecutionPhaseConstants[NodeExecutionPhase.FAILING].nodeColor}
    ${NodeExecutionPhase.SUCCEEDED} | ${nodeExecutionPhaseConstants[NodeExecutionPhase.SUCCEEDED].nodeColor}
    ${NodeExecutionPhase.ABORTED}   | ${nodeExecutionPhaseConstants[NodeExecutionPhase.ABORTED].nodeColor}
    ${NodeExecutionPhase.RUNNING}   | ${nodeExecutionPhaseConstants[NodeExecutionPhase.RUNNING].nodeColor}
    ${NodeExecutionPhase.QUEUED}    | ${nodeExecutionPhaseConstants[NodeExecutionPhase.QUEUED].nodeColor}
    ${NodeExecutionPhase.PAUSED}    | ${nodeExecutionPhaseConstants[NodeExecutionPhase.PAUSED].nodeColor}
    ${NodeExecutionPhase.UNDEFINED} | ${nodeExecutionPhaseConstants[NodeExecutionPhase.UNDEFINED].nodeColor}
  `('for each case', ({ nodeExecutionStatus, expected }) => {
    it(`should return ${expected} when called with nodeExecutionStatus = ${nodeExecutionStatus}`, () => {
      const result = getStatusColor(nodeExecutionStatus);
      expect(result).toEqual(expected);
    });
  });
});
