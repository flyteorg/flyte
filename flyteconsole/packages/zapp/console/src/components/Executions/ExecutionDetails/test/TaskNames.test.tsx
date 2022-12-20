import * as React from 'react';
import { render } from '@testing-library/react';
import { dTypes } from 'models/Graph/types';
import { TaskNames } from '../Timeline/TaskNames';

const onToggle = jest.fn();
const onAction = jest.fn();

const node1 = {
  id: 'n1',
  scopedId: 'n1',
  type: dTypes.staticNode,
  name: 'node1',
  nodes: [],
  edges: [],
};

const node2 = {
  id: 'n2',
  scopedId: 'n2',
  type: dTypes.gateNode,
  name: 'node2',
  nodes: [],
  edges: [],
};

describe('ExecutionDetails > Timeline > TaskNames', () => {
  const renderComponent = (props) => render(<TaskNames {...props} />);

  it('should render task names list', () => {
    const nodes = [node1, node2];
    const { getAllByTestId } = renderComponent({ nodes, onToggle });
    expect(getAllByTestId('task-name-item').length).toEqual(nodes.length);
  });

  it('should render task names list with resume buttons if onAction prop is passed', () => {
    const nodes = [node1, node2];
    const { getAllByTestId, getAllByTitle } = renderComponent({ nodes, onToggle, onAction });
    expect(getAllByTestId('task-name-item').length).toEqual(nodes.length);
    expect(getAllByTitle('Resume').length).toEqual(nodes.length);
  });

  it('should render task names list with expanders if nodes contain nested nodes list', () => {
    const nestedNodes = [
      { id: 't1', scopedId: 'n1', type: dTypes.task, name: 'task1', nodes: [], edges: [] },
    ];
    const nodes = [
      { ...node1, nodes: nestedNodes },
      { ...node2, nodes: nestedNodes },
    ];
    const { getAllByTestId, getAllByTitle } = renderComponent({ nodes, onToggle });
    expect(getAllByTestId('task-name-item').length).toEqual(nodes.length);
    expect(getAllByTitle('Expand row').length).toEqual(nodes.length);
  });
});
