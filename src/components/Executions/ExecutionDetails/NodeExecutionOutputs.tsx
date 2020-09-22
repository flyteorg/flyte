import { WaitForData } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { useNodeExecutionData } from 'components/hooks';
import { LiteralMapViewer, RemoteLiteralMapViewer } from 'components/Literals';
import { NodeExecution } from 'models';
import * as React from 'react';

/** Fetches and renders the output data for a given `NodeExecution` */
export const NodeExecutionOutputs: React.FC<{ execution: NodeExecution }> = ({
    execution
}) => {
    const commonStyles = useCommonStyles();
    const executionData = useNodeExecutionData(execution.id);
    return (
        <WaitForData {...executionData}>
            {() => (
                <>
                    <div className={commonStyles.detailsPanelCard}>
                        <div className={commonStyles.detailsPanelCardContent}>
                            <RemoteLiteralMapViewer
                                map={executionData.value.fullOutputs}
                                blob={executionData.value.outputs}
                            />
                        </div>
                    </div>
                </>
            )}
        </WaitForData>
    );
};
