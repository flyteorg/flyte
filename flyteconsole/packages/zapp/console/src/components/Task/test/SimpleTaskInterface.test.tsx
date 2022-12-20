import { render } from '@testing-library/react';
import { set } from 'lodash';
import { SimpleType, TypedInterface, Variable } from 'models/Common/types';
import { Task } from 'models/Task/types';
import * as React from 'react';
import { SimpleTaskInterface } from '../SimpleTaskInterface';

function setTaskInterface(task: Task, values: TypedInterface): Task {
  return set(task, 'closure.compiledTask.template.interface', values);
}

describe('SimpleTaskInterface', () => {
  const values: Record<string, Variable> = {
    value2: { type: { simple: SimpleType.INTEGER } },
    value1: { type: { simple: SimpleType.INTEGER } },
  };

  it('renders sorted inputs', () => {
    const task = setTaskInterface({} as Task, {
      inputs: { variables: { ...values } },
    });

    const { getAllByText } = render(<SimpleTaskInterface task={task as Task} />);
    const labels = getAllByText(/value/);
    expect(labels.length).toBe(2);
    expect(labels[0]).toHaveTextContent(/value1/);
    expect(labels[1]).toHaveTextContent(/value2/);
  });

  it('renders sorted outputs', () => {
    const task = setTaskInterface({} as Task, {
      outputs: { variables: { ...values } },
    });

    const { getAllByText } = render(<SimpleTaskInterface task={task as Task} />);
    const labels = getAllByText(/value/);
    expect(labels.length).toBe(2);
    expect(labels[0]).toHaveTextContent(/value1/);
    expect(labels[1]).toHaveTextContent(/value2/);
  });
});
