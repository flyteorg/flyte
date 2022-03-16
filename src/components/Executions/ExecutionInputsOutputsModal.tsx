import { Dialog, DialogContent, Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { ClosableDialogTitle } from 'components/common/ClosableDialogTitle';
import { WaitForData } from 'components/common/WaitForData';
import { LiteralMapViewer } from 'components/Literals/LiteralMapViewer';
import { RemoteLiteralMapViewer } from 'components/Literals/RemoteLiteralMapViewer';
import { emptyLiteralMapBlob } from 'models/Common/constants';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { useWorkflowExecutionData } from './useWorkflowExecution';

const useStyles = makeStyles((theme: Theme) => ({
  content: {
    paddingTop: theme.spacing(2),
  },
  dialog: {
    // Attempt to fill the window up to a max width/height
    // This will normally result in the dialog being maxWidth X maxHeight
    // except when the viewport is smaller, in which case we will take as
    // much room as possible while leaving consistent margins (enforced
    // by the MUI Dialog component)
    maxWidth: `calc(100% - ${theme.spacing(12)}px)`,
    maxHeight: `calc(100% - ${theme.spacing(12)}px)`,
    height: theme.spacing(90),
    width: theme.spacing(100),
  },
  tabs: {
    padding: `0 ${theme.spacing(2)}px`,
  },
  title: {
    display: 'flex',
    flexDirection: 'row',
  },
}));

enum TabId {
  INPUTS = 'inputs',
  OUTPUTS = 'outputs',
}

const RemoteExecutionInputs: React.FC<{ execution: Execution }> = ({ execution }) => {
  const executionData = useWorkflowExecutionData(execution.id);
  return (
    <WaitForData {...executionData} spinnerVariant="none">
      {() => (
        <RemoteLiteralMapViewer
          map={executionData.value.fullInputs}
          blob={executionData.value.inputs}
        />
      )}
    </WaitForData>
  );
};

const RemoteExecutionOutputs: React.FC<{ execution: Execution }> = ({ execution }) => {
  const executionData = useWorkflowExecutionData(execution.id);
  return (
    <WaitForData {...executionData} spinnerVariant="none">
      {() => (
        <RemoteLiteralMapViewer
          map={executionData.value.fullOutputs}
          blob={executionData.value.outputs}
        />
      )}
    </WaitForData>
  );
};

const RenderInputs: React.FC<{ execution: Execution }> = ({ execution }) => {
  // computedInputs is deprecated, but older data may still use that field.
  // For new records, the inputs will always need to be fetched separately
  return execution.closure.computedInputs ? (
    <LiteralMapViewer map={execution.closure.computedInputs} />
  ) : (
    <RemoteExecutionInputs execution={execution} />
  );
};

const RenderOutputs: React.FC<{ execution: Execution }> = ({ execution }) => {
  const outputs = execution.closure.outputs || emptyLiteralMapBlob;

  // Small outputs will be stored directly in the execution.
  // For larger outputs, we need to fetch them using the /data endpoint
  return outputs.uri ? (
    <RemoteExecutionOutputs execution={execution} />
  ) : (
    <LiteralMapViewer map={outputs.values} />
  );
};

interface RenderDialogProps {
  execution: Execution;
  onClose: () => void;
}

const RenderDialog: React.FC<RenderDialogProps> = ({ execution, onClose }) => {
  const styles = useStyles();
  const [selectedTab, setSelectedTab] = React.useState(TabId.INPUTS);
  const handleTabChange = (event: React.ChangeEvent<{}>, tabId: TabId) => setSelectedTab(tabId);
  return (
    <>
      <ClosableDialogTitle onClose={onClose}>{execution.id.name}</ClosableDialogTitle>
      <Tabs className={styles.tabs} onChange={handleTabChange} value={selectedTab}>
        <Tab value={TabId.INPUTS} label="Inputs" />
        <Tab value={TabId.OUTPUTS} label="Outputs" />
      </Tabs>
      <DialogContent className={styles.content}>
        {selectedTab === TabId.INPUTS && <RenderInputs execution={execution} />}
        {selectedTab === TabId.OUTPUTS && <RenderOutputs execution={execution} />}
      </DialogContent>
    </>
  );
};

interface ExecutionInputsOutputsModalProps {
  execution: Execution | null;
  onClose(): void;
}

/** Renders a Modal that will load/display the inputs/outputs for a given
 * Execution in a tabbed/scrollable container
 */
export const ExecutionInputsOutputsModal: React.FC<ExecutionInputsOutputsModalProps> = ({
  execution,
  onClose,
}) => {
  const styles = useStyles();
  return (
    <Dialog
      PaperProps={{ className: styles.dialog }}
      maxWidth={false}
      open={!!execution}
      onClose={onClose}
    >
      {execution ? <RenderDialog execution={execution} onClose={onClose} /> : <div />}
    </Dialog>
  );
};
