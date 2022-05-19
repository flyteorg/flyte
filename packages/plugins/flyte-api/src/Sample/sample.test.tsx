import * as React from 'react';
import { render, screen } from '@testing-library/react';
import { SampleComponent } from './index';

describe('add function', () => {
  it('SampleComponent is rendered contains correct text', () => {
    render(<SampleComponent />);
    const text = screen.getByText('Sample Text');
    expect(text).toBeInTheDocument();
  });
});
