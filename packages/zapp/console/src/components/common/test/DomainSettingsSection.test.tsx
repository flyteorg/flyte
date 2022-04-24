import { render } from '@testing-library/react';
import * as React from 'react';
import { LocalCacheProvider } from 'basics/LocalCache/ContextProvider';
import { DomainSettingsSection } from '../DomainSettingsSection';

const serviceAccount = 'default';
const rawData = 'cliOutputLocationPrefix';
const maxParallelism = 10;

const mockConfigData = {
  maxParallelism: maxParallelism,
  securityContext: { runAs: { k8sServiceAccount: serviceAccount } },
  rawOutputDataConfig: { outputLocationPrefix: rawData },
  annotations: { values: { cliAnnotationKey: 'cliAnnotationValue' } },
  labels: { values: { cliLabelKey: 'cliLabelValue' } },
};

const mockConfigDataWithoutLabels = {
  maxParallelism: maxParallelism,
  securityContext: { runAs: { k8sServiceAccount: serviceAccount } },
  rawOutputDataConfig: { outputLocationPrefix: rawData },
  annotations: { values: { cliAnnotationKey: 'cliAnnotationValue' } },
};

const mockConfigDataWithoutLabelsAndAnnotations = {
  maxParallelism: maxParallelism,
  securityContext: { runAs: { k8sServiceAccount: serviceAccount } },
  rawOutputDataConfig: { outputLocationPrefix: rawData },
};

describe('DomainSettingsSection', () => {
  it('should not render a block if config data passed is empty', () => {
    const { container } = render(
      <LocalCacheProvider>
        <DomainSettingsSection configData={{}} />
      </LocalCacheProvider>,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it('should render a section without IAMRole data', () => {
    const { queryByText, queryAllByRole } = render(
      <LocalCacheProvider>
        <DomainSettingsSection configData={mockConfigData} />
      </LocalCacheProvider>,
    );
    expect(queryByText('Domain Settings')).toBeInTheDocument();
    // should display serviceAccount value
    expect(queryByText(serviceAccount)).toBeInTheDocument();
    // should display rawData value
    expect(queryByText(rawData)).toBeInTheDocument();
    // should display maxParallelism value
    expect(queryByText(maxParallelism)).toBeInTheDocument();
    // should display 2 data tables
    const tables = queryAllByRole('table');
    expect(tables).toHaveLength(2);
    // should display a placeholder text, as role was not passed
    const emptyRole = queryByText('Inherits from project level values');
    expect(emptyRole).toBeInTheDocument();
  });

  it('should render a section without IAMRole and Labels data', () => {
    const { queryByText, queryAllByText, queryAllByRole } = render(
      <LocalCacheProvider>
        <DomainSettingsSection configData={mockConfigDataWithoutLabels} />
      </LocalCacheProvider>,
    );
    expect(queryByText('Domain Settings')).toBeInTheDocument();
    // should display serviceAccount value
    expect(queryByText(serviceAccount)).toBeInTheDocument();
    // should display rawData value
    expect(queryByText(rawData)).toBeInTheDocument();
    // should display maxParallelism value
    expect(queryByText(maxParallelism)).toBeInTheDocument();
    // should display 1 data table
    const tables = queryAllByRole('table');
    expect(tables).toHaveLength(1);
    // should display two placeholder text, as role and labels were not passed
    const inheritedPlaceholders = queryAllByText('Inherits from project level values');
    expect(inheritedPlaceholders).toHaveLength(2);
  });

  it('should render a section without IAMRole, Labels, Annotations data', () => {
    const { queryByText, queryAllByText, queryByRole } = render(
      <LocalCacheProvider>
        <DomainSettingsSection configData={mockConfigDataWithoutLabelsAndAnnotations} />
      </LocalCacheProvider>,
    );
    expect(queryByText('Domain Settings')).toBeInTheDocument();
    // should display serviceAccount value
    expect(queryByText(serviceAccount)).toBeInTheDocument();
    // should display rawData value
    expect(queryByText(rawData)).toBeInTheDocument();
    // should display maxParallelism value
    expect(queryByText(maxParallelism)).toBeInTheDocument();
    // should not display any data tables
    const tables = queryByRole('table');
    expect(tables).not.toBeInTheDocument();
    // should display three placeholder text, as role, labels, annotations were not passed
    const inheritedPlaceholders = queryAllByText('Inherits from project level values');
    expect(inheritedPlaceholders).toHaveLength(3);
  });
});
