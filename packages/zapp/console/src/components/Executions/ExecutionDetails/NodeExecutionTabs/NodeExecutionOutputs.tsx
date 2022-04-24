import { PanelSection } from 'components/common/PanelSection';
import { WaitForData } from 'components/common/WaitForData';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { LiteralMapViewer } from 'components/Literals/LiteralMapViewer';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';

/** Fetches and renders the output data for a given `NodeExecution` */
export const NodeExecutionOutputs: React.FC<{ execution: NodeExecution }> = ({ execution }) => {
  const executionData = useNodeExecutionData(execution.id);
  return (
    <WaitForData {...executionData}>
      <PanelSection>
        <LiteralMapViewer map={executionData.value.fullOutputs} />
      </PanelSection>
    </WaitForData>
  );
};
