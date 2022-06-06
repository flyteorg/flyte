import * as React from 'react';
import { render, screen } from '@testing-library/react';
import { SampleComponent, add } from './index';

describe('add function', () => {
  it('should add two number together', () => {
    const result = add(10, 5);
    expect(result).toBe(15);
  });

  it('NavBar contains correct text', () => {
    render(<SampleComponent />);
    const text = screen.getByText("It's me - Navigation Bar");
    expect(text).toBeInTheDocument();
  });
});
