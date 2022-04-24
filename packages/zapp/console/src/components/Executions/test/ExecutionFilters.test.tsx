import { render } from '@testing-library/react';
import * as React from 'react';
import { ExecutionFilters, ExecutionFiltersProps } from '../ExecutionFilters';
import { FilterState } from '../filters/types';

const filterLabel1 = 'Single Filter';
const filterLabel2 = 'Multiple Filter';
const filterLabel3 = 'Search Filter';
const filters: FilterState[] = [
  {
    active: true,
    label: filterLabel1,
    type: 'single',
    button: {
      open: false,
      setOpen: jest.fn(),
      onClick: jest.fn(),
    },
    getFilter: jest.fn(),
    onReset: jest.fn(),
    onChange: jest.fn(),
  },
  {
    active: true,
    label: filterLabel2,
    type: 'multi',
    button: {
      open: false,
      setOpen: jest.fn(),
      onClick: jest.fn(),
    },
    getFilter: jest.fn(),
    onReset: jest.fn(),
    onChange: jest.fn(),
  },
  {
    active: true,
    label: filterLabel3,
    type: 'search',
    button: {
      open: false,
      setOpen: jest.fn(),
      onClick: jest.fn(),
    },
    getFilter: jest.fn(),
    onReset: jest.fn(),
    onChange: jest.fn(),
  },
];

describe('ExecutionFilters', () => {
  describe('generic hook filters', () => {
    it('should display all provided filters', () => {
      const props: ExecutionFiltersProps = {
        filters,
      };
      const { getAllByRole } = render(<ExecutionFilters {...props} />);
      const renderedFilters = getAllByRole(/button/i);
      expect(renderedFilters).toHaveLength(3);
      expect(renderedFilters[0]).toHaveTextContent(filterLabel1);
      expect(renderedFilters[1]).toHaveTextContent(filterLabel2);
      expect(renderedFilters[2]).toHaveTextContent(filterLabel3);
    });

    it('should not display hidden filters provided filters', () => {
      const hiddenFilter = { ...filters[2], hidden: true };
      const props: ExecutionFiltersProps = {
        filters: [filters[0], filters[1], hiddenFilter],
      };
      const { getAllByRole } = render(<ExecutionFilters {...props} />);
      const renderedFilters = getAllByRole(/button/i);
      expect(renderedFilters).toHaveLength(2);
      expect(renderedFilters[0]).toHaveTextContent(filterLabel1);
      expect(renderedFilters[1]).toHaveTextContent(filterLabel2);
    });
  });

  describe('clear executions button', () => {
    it('should not be rendered, when chartIds is not provided', () => {
      const props: ExecutionFiltersProps = {
        filters,
      };
      const { queryByTestId } = render(<ExecutionFilters {...props} />);
      expect(queryByTestId('clear-charts')).toBeNull();
    });

    it('should not be rendered, when chartIds is empty', () => {
      const chartIds = [];
      const props: ExecutionFiltersProps = {
        filters,
        chartIds,
        clearCharts: jest.fn(),
      };
      const { queryByTestId } = render(<ExecutionFilters {...props} />);
      expect(queryByTestId('clear-charts')).toBeNull();
    });

    it('should be rendered, when chartIds is not empty', () => {
      const chartIds = ['id'];
      const props: ExecutionFiltersProps = {
        filters,
        chartIds,
        clearCharts: jest.fn(),
      };
      const { queryByTestId } = render(<ExecutionFilters {...props} />);
      expect(queryByTestId('clear-charts')).toBeDefined();
    });
  });

  describe('custom hook filters', () => {
    it('should display onlyMyExecution checkbox when corresponding filter state was provided', () => {
      const props: ExecutionFiltersProps = {
        filters: [],
        onlyMyExecutionsFilterState: {
          onlyMyExecutionsValue: false,
          isFilterDisabled: false,
          onOnlyMyExecutionsFilterChange: jest.fn(),
        },
      };
      const { getAllByRole } = render(<ExecutionFilters {...props} />);
      const checkboxes = getAllByRole(/checkbox/i) as HTMLInputElement[];
      expect(checkboxes).toHaveLength(1);
      expect(checkboxes[0]).toBeTruthy();
      expect(checkboxes[0]).toBeEnabled();
    });

    it('should display showArchived checkbox when corresponding props were provided', () => {
      const props: ExecutionFiltersProps = {
        filters: [],
        showArchived: true,
        onArchiveFilterChange: jest.fn(),
      };
      const { getAllByRole } = render(<ExecutionFilters {...props} />);
      const checkboxes = getAllByRole(/checkbox/i) as HTMLInputElement[];
      expect(checkboxes).toHaveLength(1);
      expect(checkboxes[0]).toBeTruthy();
      expect(checkboxes[0]).toBeEnabled();
    });

    it('should display 2 checkbox filters', () => {
      const props: ExecutionFiltersProps = {
        filters: [],
        onlyMyExecutionsFilterState: {
          onlyMyExecutionsValue: false,
          isFilterDisabled: false,
          onOnlyMyExecutionsFilterChange: jest.fn(),
        },
        showArchived: true,
        onArchiveFilterChange: jest.fn(),
      };
      const { getAllByRole } = render(<ExecutionFilters {...props} />);
      const checkboxes = getAllByRole(/checkbox/i) as HTMLInputElement[];
      expect(checkboxes).toHaveLength(2);
      expect(checkboxes[0]).toBeTruthy();
      expect(checkboxes[0]).toBeEnabled();
      expect(checkboxes[1]).toBeTruthy();
      expect(checkboxes[1]).toBeEnabled();
    });
  });
});
