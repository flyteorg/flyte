import { act, render } from '@testing-library/react';
import {
    createMockLaunchPlan,
    mockLaunchPlanSchedules
} from 'models/__mocks__/launchPlanData';
import { FilterOperation, FilterOperationName } from 'models/AdminEntity/types';
import { NamedEntityIdentifier } from 'models/Common/types';
import { listLaunchPlans } from 'models/Launch/api';
import { LaunchPlan, LaunchPlanState } from 'models/Launch/types';
import * as React from 'react';
import { WorkflowSchedules } from '../WorkflowSchedules';

jest.mock('models/Launch/api');

describe('WorkflowSchedules', () => {
    const mockListLaunchPlans = listLaunchPlans as jest.Mock<
        ReturnType<typeof listLaunchPlans>
    >;
    const workflowId: NamedEntityIdentifier = {
        project: 'project',
        domain: 'domain',
        name: 'name'
    };
    let launchPlans: LaunchPlan[];

    const renderSchedules = async () => {
        await act(() => {
            render(<WorkflowSchedules workflowId={workflowId} />);
            return new Promise(resolve => setTimeout(resolve, 0));
        });
    };

    beforeEach(() => {
        launchPlans = [
            createMockLaunchPlan('EveryTenMinutes', 'abcdefg'),
            createMockLaunchPlan('Daily6AM', 'abcdefg')
        ];
        launchPlans[0].spec.entityMetadata.schedule =
            mockLaunchPlanSchedules.everyTenMinutes;
        launchPlans[1].spec.entityMetadata.schedule =
            mockLaunchPlanSchedules.everyDay6AM;
        mockListLaunchPlans.mockResolvedValue({ entities: launchPlans });
    });

    it('should only request active schedules', async () => {
        const expectedFilterOpration: FilterOperation = {
            key: 'state',
            operation: FilterOperationName.EQ,
            value: LaunchPlanState.ACTIVE
        };

        await renderSchedules();
        expect(mockListLaunchPlans).toHaveBeenCalledWith(
            expect.any(Object),
            expect.objectContaining({
                filter: expect.arrayContaining([expectedFilterOpration])
            })
        );
    });

    it('should request schedules for only the given workflow Id', async () => {
        const expectedFilterOperations: FilterOperation[] = [
            {
                key: 'workflow.name',
                operation: FilterOperationName.EQ,
                value: workflowId.name
            },
            {
                key: 'workflow.domain',
                operation: FilterOperationName.EQ,
                value: workflowId.domain
            },
            {
                key: 'workflow.project',
                operation: FilterOperationName.EQ,
                value: workflowId.project
            }
        ];
        await renderSchedules();
        expect(mockListLaunchPlans).toHaveBeenCalledWith(
            expect.any(Object),
            expect.objectContaining({
                filter: expect.arrayContaining(expectedFilterOperations)
            })
        );
    });
});
