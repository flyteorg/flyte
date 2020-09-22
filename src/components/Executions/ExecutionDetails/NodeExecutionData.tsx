import { Card, CardContent } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import { WaitForData } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { useNodeExecutionData } from 'components/hooks';
import { LiteralMapViewer, RemoteLiteralMapViewer } from 'components/Literals';
import { NodeExecution } from 'models';
import * as React from 'react';

/** Fetches and renders the execution data (inputs/outputs for a given
 * `NodeExecution`) */
export const NodeExecutionData: React.FC<{ execution: NodeExecution }> = ({
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
                            <header>
                                <Typography variant="subtitle2">
                                    Inputs
                                </Typography>
                            </header>
                            <section>
                                <RemoteLiteralMapViewer
                                    map={executionData.value.fullInputs}
                                    blob={executionData.value.inputs}
                                />
                            </section>
                        </div>
                    </div>
                    <div className={commonStyles.detailsPanelCard}>
                        <div className={commonStyles.detailsPanelCardContent}>
                            <header>
                                <Typography variant="subtitle2">
                                    Outputs
                                </Typography>
                            </header>
                            <section>
                                <RemoteLiteralMapViewer
                                    map={executionData.value.fullOutputs}
                                    blob={executionData.value.outputs}
                                />
                            </section>
                        </div>
                    </div>
                </>
            )}
        </WaitForData>
    );
};
