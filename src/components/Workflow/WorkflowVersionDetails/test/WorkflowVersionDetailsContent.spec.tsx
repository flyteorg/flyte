import * as React from 'react';
import * as renderer from 'react-test-renderer';

jest.mock('components/flytegraph/Graph.tsx', () => {
    const React = require('react');
    return {
        // Test environment doesn't have support for canvas, so we need
        // to mock away the flytegraph components.
        Graph: () => <div id="workflow-graph" />
    };
});

jest.mock('components/common/DetailsPanelContent.tsx', () => {
    const React = require('react');
    return {
        // Portals aren't working correctly, so we'll use a dummy version of
        // DetailsPanel that renders inline
        DetailsPanelContent: ({ children }: React.Props<{}>) => (
            <div id="details-panel">{children}</div>
        )
    };
});

import { Workflow } from 'models';
import {
    createMockWorkflow,
    createMockWorkflowClosure
} from 'models/__mocks__/workflowData';

import {
    WorkflowVersionDetailsContent,
    WorkflowVersionDetailsContentProps
} from '../WorkflowVersionDetailsContent';

describe('WorkflowVersionDetailsContent', () => {
    let props: WorkflowVersionDetailsContentProps;
    let workflow: Workflow;
    beforeEach(() => {
        workflow = createMockWorkflow('my_workflow');
        workflow.closure = createMockWorkflowClosure();
        props = {
            workflow
        };
    });

    const renderSnapshot = () =>
        renderer.create(<WorkflowVersionDetailsContent {...props} />).toJSON();

    it('should match known snapshot', () => {
        expect(renderSnapshot()).toMatchSnapshot();
    });
});
