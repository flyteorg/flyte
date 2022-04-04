import { PanelSection } from 'components/common/PanelSection';
import { WaitForData } from 'components/common/WaitForData';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { RemoteLiteralMapViewer } from 'components/Literals/RemoteLiteralMapViewer';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';

/** Fetches and renders the input data for a given `NodeExecution` */
export const NodeExecutionInputs: React.FC<{ execution: NodeExecution }> = ({ execution }) => {
  const executionData = useNodeExecutionData(execution.id);
  return (
    <WaitForData {...executionData}>
      <PanelSection>
        <RemoteLiteralMapViewer
          map={executionData.value.fullInputs}
          blob={executionData.value.inputs}
        />
      </PanelSection>
    </WaitForData>
  );
};
