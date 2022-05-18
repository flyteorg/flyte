import { render } from '@testing-library/react';
import { unknownValueString } from 'common/constants';
import * as React from 'react';
import { long } from 'test/utils';
import { TaskExecutionDetails } from '../TaskExecutionDetails';

const date = { seconds: long(5), nanos: 0 };
const duration = { seconds: long(0), nanos: 0 };

const dateContent = '1/1/1970 12:00:05 AM UTC (52 years ago)';

describe('TaskExecutionDetails', () => {
  it('should render details with task started info and duration', () => {
    const { queryByText } = render(<TaskExecutionDetails startedAt={date} duration={duration} />);

    expect(queryByText('started')).toBeInTheDocument();
    expect(queryByText('last updated')).not.toBeInTheDocument();
    expect(queryByText(dateContent)).toBeInTheDocument();
    expect(queryByText('run time')).toBeInTheDocument();
    expect(queryByText('0s')).toBeInTheDocument();
  });

  it('should render details with task started info without duration', () => {
    const { queryByText } = render(<TaskExecutionDetails startedAt={date} />);

    expect(queryByText('started')).toBeInTheDocument();
    expect(queryByText('last updated')).not.toBeInTheDocument();
    expect(queryByText(dateContent)).toBeInTheDocument();
    expect(queryByText('run time')).toBeInTheDocument();
    expect(queryByText(unknownValueString)).toBeInTheDocument();
  });

  it('should render details with task updated info and duration', () => {
    const { queryByText } = render(<TaskExecutionDetails updatedAt={date} duration={duration} />);

    expect(queryByText('started')).not.toBeInTheDocument();
    expect(queryByText('last updated')).toBeInTheDocument();
    expect(queryByText(dateContent)).toBeInTheDocument();
    expect(queryByText('run time')).not.toBeInTheDocument();
    expect(queryByText('0s')).not.toBeInTheDocument();
  });

  it('should render details with task updated info without duration', () => {
    const { queryByText } = render(<TaskExecutionDetails updatedAt={date} />);

    expect(queryByText('started')).not.toBeInTheDocument();
    expect(queryByText('last updated')).toBeInTheDocument();
    expect(queryByText(dateContent)).toBeInTheDocument();
    expect(queryByText('run time')).not.toBeInTheDocument();
    expect(queryByText(unknownValueString)).not.toBeInTheDocument();
  });
});
