import { Button, Dialog, IconButton } from '@material-ui/core';
import * as React from 'react';
import { ResourceIdentifier, Identifier } from 'models/Common/types';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { getTask } from 'models/Task/api';
import { LaunchFormDialog } from 'components/Launch/LaunchForm/LaunchFormDialog';
import { NodeExecutionIdentifier } from 'models/Execution/types';
import { useNodeExecution, useNodeExecutionData } from 'components/hooks/useNodeExecution';
import { literalsToLiteralValueMap } from 'components/Launch/LaunchForm/utils';
import { TaskInitialLaunchParameters } from 'components/Launch/LaunchForm/types';
import { NodeExecutionPhase } from 'models/Execution/enums';
import Close from '@material-ui/icons/Close';
import { useEffect, useState } from 'react';
import { NodeExecutionDetails } from '../types';
import t from './strings';
import { ExecutionNodeDeck } from './ExecutionNodeDeck';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';

const useStyles = makeStyles((theme: Theme) => {
  return {
    actionsContainer: {
      borderTop: `1px solid ${theme.palette.divider}`,
      marginTop: theme.spacing(2),
      paddingTop: theme.spacing(2),
      '& button': {
        marginRight: theme.spacing(1),
      },
    },
    dialog: {
      maxWidth: `calc(100% - ${theme.spacing(12)}px)`,
      maxHeight: `calc(100% - ${theme.spacing(12)}px)`,
      height: theme.spacing(90),
      width: theme.spacing(110),
      '& iframe': {
        border: 'none',
      },
    },
    dialogTitle: {
      display: 'flex',
      alignItems: 'center',
      padding: theme.spacing(2),
      paddingBottom: theme.spacing(0),
      fontFamily: 'Open sans',
    },
    deckTitle: {
      flexGrow: 1,
      textAlign: 'center',
      fontSize: '24px',
      lineHeight: '32px',
      marginBlock: 0,
      paddingTop: theme.spacing(2),
      paddingBottom: theme.spacing(2),
    },
    close: {
      position: 'absolute',
      right: theme.spacing(2),
    },
  };
});

interface ExecutionDetailsActionsProps {
  details?: NodeExecutionDetails;
  nodeExecutionId: NodeExecutionIdentifier;
  phase: NodeExecutionPhase;
}

export const ExecutionDetailsActions = ({
  details,
  nodeExecutionId,
  phase,
}: ExecutionDetailsActionsProps): JSX.Element => {
  const styles = useStyles();

  const [showLaunchForm, setShowLaunchForm] = useState<boolean>(false);
  const [showResumeForm, setShowResumeForm] = useState<boolean>(false);

  const [initialParameters, setInitialParameters] = useState<
    TaskInitialLaunchParameters | undefined
  >(undefined);
  const executionData = useNodeExecutionData(nodeExecutionId);
  const execution = useNodeExecution(nodeExecutionId);
  const { compiledWorkflowClosure } = useNodeExecutionContext();
  const id = details?.taskTemplate?.id;
  const compiledNode = (compiledWorkflowClosure?.primary.template.nodes ?? []).find(
    (node) => node.id === nodeExecutionId.nodeId,
  );

  useEffect(() => {
    if (!id) {
      return;
    }

    (async () => {
      const task = await getTask(id!);

      const literals = executionData.value.fullInputs?.literals;
      const taskInputsTypes = task.closure.compiledTask.template?.interface?.inputs?.variables;

      const tempInitialParameters: TaskInitialLaunchParameters = {
        values: literals && taskInputsTypes && literalsToLiteralValueMap(literals, taskInputsTypes),
        taskId: id as Identifier | undefined,
      };

      setInitialParameters(tempInitialParameters);
    })();
  }, [details]);

  const [showDeck, setShowDeck] = React.useState(false);
  const onCloseDeck = () => setShowDeck(false);

  const rerunOnClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setShowLaunchForm(true);
  };

  const onResumeClick = (e: React.MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    setShowResumeForm(true);
  };

  return (
    <>
      <div className={styles.actionsContainer}>
        {execution?.value?.closure?.deckUri && (
          <Button
            variant="outlined"
            color="primary"
            onClick={() => setShowDeck(true)}
            disabled={phase !== NodeExecutionPhase.SUCCEEDED}
          >
            {t('flyteDeck')}
          </Button>
        )}
        {id && initialParameters && details && (
          <Button variant="outlined" color="primary" onClick={rerunOnClick}>
            {t('rerun')}
          </Button>
        )}
        {phase === NodeExecutionPhase.PAUSED && (
          <Button variant="outlined" color="primary" onClick={onResumeClick}>
            {t('resume')}
          </Button>
        )}
      </div>
      {id && initialParameters && (
        <LaunchFormDialog
          id={id as ResourceIdentifier}
          initialParameters={initialParameters}
          showLaunchForm={showLaunchForm}
          setShowLaunchForm={setShowLaunchForm}
        />
      )}
      {compiledNode && (
        <LaunchFormDialog
          compiledNode={compiledNode}
          initialParameters={initialParameters}
          nodeId={nodeExecutionId.nodeId}
          showLaunchForm={showResumeForm}
          setShowLaunchForm={setShowResumeForm}
        />
      )}
      {execution?.value?.closure?.deckUri && (
        <Dialog PaperProps={{ className: styles.dialog }} maxWidth={false} open={showDeck}>
          <div className={styles.dialogTitle}>
            <h2 className={styles.deckTitle}>{t('flyteDeck')}</h2>
            <IconButton aria-label="close" onClick={onCloseDeck} className={styles.close}>
              <Close />
            </IconButton>
          </div>
          <ExecutionNodeDeck nodeExecutionId={nodeExecutionId} />
        </Dialog>
      )}
    </>
  );
};
