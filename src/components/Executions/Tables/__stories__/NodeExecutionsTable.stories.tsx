import { makeStyles, Theme } from '@material-ui/core/styles';
import { storiesOf } from '@storybook/react';
import { NodeExecutionDetailsContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { makeNodeExecutionListQuery } from 'components/Executions/nodeExecutionQueries';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import * as React from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { NodeExecutionsTable } from '../NodeExecutionsTable';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    borderLeft: `1px solid ${theme.palette.grey[400]}`,
    display: 'flex',
    height: '100vh',
    padding: `${theme.spacing(2)}px 0`,
    width: '100vw'
  }
}));

const fixture = basicPythonWorkflow.generate();
const workflowExecution = fixture.workflowExecutions.top.data;
const workflowId = { ...fixture.workflowExecutions.top.data.id, version: '0.1' };
const compiledWorkflowClosure = null;

const getNodeExecutionDetails = async () => {
  return {
    displayId: 'node0',
    displayName: 'basic.byton.workflow.unique.task_name',
    displayType: 'Python-Task'
  };
};

const stories = storiesOf('Tables/NodeExecutionsTable', module);
stories.addDecorator(story => {
  return <div className={useStyles().container}>{story()}</div>;
});
stories.add('Basic', () => {
  const query = useQuery(makeNodeExecutionListQuery(useQueryClient(), workflowExecution.id));
  return query.data ? (
    <NodeExecutionDetailsContext.Provider value={{ getNodeExecutionDetails, workflowId, compiledWorkflowClosure }}>
      <NodeExecutionsTable nodeExecutions={query.data} />
    </NodeExecutionDetailsContext.Provider>
  ) : (
    <div />
  );
});
stories.add('With no items', () => {
  return (
    <NodeExecutionDetailsContext.Provider value={{ getNodeExecutionDetails, workflowId, compiledWorkflowClosure }}>
      <NodeExecutionsTable nodeExecutions={[]} />
    </NodeExecutionDetailsContext.Provider>
  );
});
