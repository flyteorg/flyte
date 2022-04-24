import { render, waitFor } from '@testing-library/react';
import { createMockLaunchPlan, mockLaunchPlanSchedules } from 'models/__mocks__/launchPlanData';
import { FilterOperation, FilterOperationName } from 'models/AdminEntity/types';
import { ResourceIdentifier, ResourceType } from 'models/Common/types';
import { listLaunchPlans } from 'models/Launch/api';
import { LaunchPlan, LaunchPlanState } from 'models/Launch/types';
import * as React from 'react';
import { EntitySchedules } from '../EntitySchedules';
import t from '../strings';

jest.mock('models/Launch/api');

describe('EntitySchedules', () => {
  const mockListLaunchPlans = listLaunchPlans as jest.Mock<ReturnType<typeof listLaunchPlans>>;
  const id: ResourceIdentifier = {
    resourceType: ResourceType.WORKFLOW,
    project: 'project',
    domain: 'domain',
    name: 'name',
  };
  let launchPlans: LaunchPlan[];

  const renderSchedules = async () => {
    const result = render(<EntitySchedules id={id} />);
    await waitFor(() => result.getByText(t('schedulesHeader')));
    return result;
  };

  beforeEach(() => {
    launchPlans = [
      createMockLaunchPlan('EveryTenMinutes', 'abcdefg'),
      createMockLaunchPlan('Daily6AM', 'abcdefg'),
    ];
    launchPlans[0].spec.entityMetadata.schedule = mockLaunchPlanSchedules.everyTenMinutes;
    launchPlans[1].spec.entityMetadata.schedule = mockLaunchPlanSchedules.everyDay6AM;
    mockListLaunchPlans.mockResolvedValue({ entities: launchPlans });
  });

  it('should only request active schedules', async () => {
    const expectedFilterOpration: FilterOperation = {
      key: 'state',
      operation: FilterOperationName.EQ,
      value: LaunchPlanState.ACTIVE,
    };

    await renderSchedules();
    expect(mockListLaunchPlans).toHaveBeenCalledWith(
      expect.any(Object),
      expect.objectContaining({
        filter: expect.arrayContaining([expectedFilterOpration]),
      }),
    );
  });

  it('should request schedules for only the given workflow Id', async () => {
    const expectedFilterOperations: FilterOperation[] = [
      {
        key: 'workflow.name',
        operation: FilterOperationName.EQ,
        value: id.name,
      },
      {
        key: 'workflow.domain',
        operation: FilterOperationName.EQ,
        value: id.domain,
      },
      {
        key: 'workflow.project',
        operation: FilterOperationName.EQ,
        value: id.project,
      },
    ];
    await renderSchedules();
    expect(mockListLaunchPlans).toHaveBeenCalledWith(
      expect.any(Object),
      expect.objectContaining({
        filter: expect.arrayContaining(expectedFilterOperations),
      }),
    );
  });
});
