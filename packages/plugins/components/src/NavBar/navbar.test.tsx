import * as React from 'react';
import { render, screen } from '@testing-library/react';
import { NavBar, add } from './index';

describe('add function', () => {
  it('should add two number together', () => {
    const result = add(10, 5);
    expect(result).toBe(15);
  });

  it('NavBar contains correct text', () => {
    render(<NavBar />);
    const text = screen.getByText('NASTYA IS HERE');
    expect(text).toBeInTheDocument();
  });
});
