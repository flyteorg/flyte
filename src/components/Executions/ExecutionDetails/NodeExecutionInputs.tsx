import { useCommonStyles } from 'components/common/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { RemoteLiteralMapViewer } from 'components/Literals/RemoteLiteralMapViewer';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';

/** Fetches and renders the input data for a given `NodeExecution` */
export const NodeExecutionInputs: React.FC<{ execution: NodeExecution }> = ({ execution }) => {
  const commonStyles = useCommonStyles();
  const executionData = useNodeExecutionData(execution.id);
  return (
    <WaitForData {...executionData}>
      {() => (
        <>
          <div className={commonStyles.detailsPanelCard}>
            <div className={commonStyles.detailsPanelCardContent}>
              <RemoteLiteralMapViewer
                map={executionData.value.fullInputs}
                blob={executionData.value.inputs}
              />
            </div>
          </div>
        </>
      )}
    </WaitForData>
  );
};
