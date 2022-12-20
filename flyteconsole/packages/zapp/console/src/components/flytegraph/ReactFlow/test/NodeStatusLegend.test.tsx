import * as React from 'react';
import { fireEvent, render } from '@testing-library/react';
import { Legend } from '../NodeStatusLegend';
import { graphNodePhasesList } from '../utils';

describe('flytegraph > ReactFlow > NodeStatusLegend', () => {
  const renderComponent = (props) => render(<Legend {...props} />);

  it('should render just the Legend button, if initialIsVisible was not passed', () => {
    const { queryByTitle, queryByTestId } = renderComponent({});
    expect(queryByTitle('Show Legend')).toBeInTheDocument();
    expect(queryByTestId('legend-table')).not.toBeInTheDocument();
  });

  it('should render just the Legend button, if initialIsVisible is false', () => {
    const { queryByTitle, queryByTestId } = renderComponent({ initialIsVisible: false });
    expect(queryByTitle('Show Legend')).toBeInTheDocument();
    expect(queryByTestId('legend-table')).not.toBeInTheDocument();
  });

  it('should render Legend table, if initialIsVisible is true', () => {
    const { queryByTitle, queryByTestId, queryAllByTestId } = renderComponent({
      initialIsVisible: true,
    });
    expect(queryByTitle('Show Legend')).not.toBeInTheDocument();
    expect(queryByTitle('Hide Legend')).toBeInTheDocument();
    expect(queryByTestId('legend-table')).toBeInTheDocument();
    // the number of items should match the graphNodePhasesList const plus one extra for nested nodes
    expect(queryAllByTestId('legend-item').length).toEqual(graphNodePhasesList.length + 1);
  });

  it('should render Legend table on button click, and hide it, when clicked again', async () => {
    const { getByRole, queryByTitle, queryByTestId } = renderComponent({});
    expect(queryByTitle('Show Legend')).toBeInTheDocument();
    expect(queryByTitle('Hide Legend')).not.toBeInTheDocument();
    expect(queryByTestId('legend-table')).not.toBeInTheDocument();

    const button = getByRole('button');
    await fireEvent.click(button);

    expect(queryByTitle('Show Legend')).not.toBeInTheDocument();
    expect(queryByTitle('Hide Legend')).toBeInTheDocument();
    expect(queryByTestId('legend-table')).toBeInTheDocument();

    await fireEvent.click(button);

    expect(queryByTitle('Show Legend')).toBeInTheDocument();
    expect(queryByTitle('Hide Legend')).not.toBeInTheDocument();
    expect(queryByTestId('legend-table')).not.toBeInTheDocument();
  });
});
