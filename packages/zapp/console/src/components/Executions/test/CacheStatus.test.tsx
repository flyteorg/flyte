import { render } from '@testing-library/react';
import { cacheStatusMessages, viewSourceExecutionString } from 'components/Executions/constants';
import { CatalogCacheStatus } from 'models/Execution/enums';
import { mockWorkflowExecutionId } from 'models/Execution/__mocks__/constants';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { Routes } from 'routes/routes';
import { CacheStatus } from '../CacheStatus';

describe('Executions > CacheStatus', () => {
  const renderComponent = (props) =>
    render(
      <MemoryRouter>
        <CacheStatus {...props} />
      </MemoryRouter>,
    );

  describe('check renders', () => {
    it('should not render anything, if cacheStatus is null', () => {
      const cacheStatus = null;
      const { container } = renderComponent({ cacheStatus });

      expect(container).toBeEmptyDOMElement();
    });

    it('should not render anything, if cacheStatus is undefined', () => {
      const cacheStatus = undefined;
      const { container } = renderComponent({ cacheStatus });

      expect(container).toBeEmptyDOMElement();
    });

    it('should render text with icon, if no variant was provided', () => {
      const cacheStatus = CatalogCacheStatus.CACHE_POPULATED;
      const { queryByText, queryByTestId } = renderComponent({ cacheStatus });

      expect(queryByText(cacheStatusMessages[cacheStatus])).toBeInTheDocument();
      expect(queryByTestId('cache-icon')).toBeInTheDocument();
    });

    it('should render text with icon, if variant = normal', () => {
      const cacheStatus = CatalogCacheStatus.CACHE_POPULATED;
      const { queryByText, queryByTestId } = renderComponent({ cacheStatus, variant: 'normal' });

      expect(queryByText(cacheStatusMessages[cacheStatus])).toBeInTheDocument();
      expect(queryByTestId('cache-icon')).toBeInTheDocument();
    });

    it('should not render text, if variant = iconOnly', () => {
      const cacheStatus = CatalogCacheStatus.CACHE_POPULATED;
      const { queryByText, queryByTestId } = renderComponent({ cacheStatus, variant: 'iconOnly' });

      expect(queryByText(cacheStatusMessages[cacheStatus])).not.toBeInTheDocument();
      expect(queryByTestId('cache-icon')).toBeInTheDocument();
    });

    it('should render source execution link for cache hits', () => {
      const cacheStatus = CatalogCacheStatus.CACHE_HIT;
      const sourceTaskExecutionId = {
        taskId: { ...mockWorkflowExecutionId, version: '1' },
        nodeExecutionId: { nodeId: 'n1', executionId: mockWorkflowExecutionId },
      };
      const { getByText } = renderComponent({ cacheStatus, sourceTaskExecutionId });
      const linkEl = getByText(viewSourceExecutionString);

      expect(linkEl.getAttribute('href')).toEqual(
        Routes.ExecutionDetails.makeUrl(mockWorkflowExecutionId),
      );
    });
  });

  describe('check cache statuses', () => {
    describe.each`
      cacheStatus                                | expected
      ${CatalogCacheStatus.CACHE_DISABLED}       | ${cacheStatusMessages[CatalogCacheStatus.CACHE_DISABLED]}
      ${CatalogCacheStatus.CACHE_HIT}            | ${cacheStatusMessages[CatalogCacheStatus.CACHE_HIT]}
      ${CatalogCacheStatus.CACHE_LOOKUP_FAILURE} | ${cacheStatusMessages[CatalogCacheStatus.CACHE_LOOKUP_FAILURE]}
      ${CatalogCacheStatus.CACHE_MISS}           | ${cacheStatusMessages[CatalogCacheStatus.CACHE_MISS]}
      ${CatalogCacheStatus.CACHE_POPULATED}      | ${cacheStatusMessages[CatalogCacheStatus.CACHE_POPULATED]}
      ${CatalogCacheStatus.CACHE_PUT_FAILURE}    | ${cacheStatusMessages[CatalogCacheStatus.CACHE_PUT_FAILURE]}
    `('for each case', ({ cacheStatus, expected }) => {
      it(`renders correct text ${expected} for status ${cacheStatus}`, async () => {
        const { queryByText } = renderComponent({ cacheStatus });

        expect(queryByText(cacheStatusMessages[cacheStatus])).toBeInTheDocument();
      });
    });
  });
});
