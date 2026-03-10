import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { GroupFolderItem } from './GroupFolderItem'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { ActionWithChildren } from '../state/types'

// Helper function to create a mock node with phase counts
const createMockNode = (
  phaseCounts: Partial<Record<ActionPhase, number>>,
): ActionWithChildren => ({
  children: [],
  groupChildren: {},
  childrenPhaseCounts: Object.fromEntries(
    Object.entries(phaseCounts)
      .filter(([_, v]) => v !== undefined)
      .map(([k, v]) => [k, v as number]),
  ),
  meetsFilter: true,
  $typeName: 'flyteidl2.workflow.EnrichedAction',
  isGroup: false,
})

describe('GroupFolderItem component', () => {
  it('shows unspecified icon when all children are UNSPECIFIED', () => {
    const node = createMockNode({ [ActionPhase.UNSPECIFIED]: 3 })
    render(
      <GroupFolderItem
        displayText="Group A"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('unspecified-folder')).toBeInTheDocument()
  })

  it('shows Succeeded when mixed with UNSPECIFIED', () => {
    const node = createMockNode({
      [ActionPhase.UNSPECIFIED]: 2,
      [ActionPhase.SUCCEEDED]: 3,
    })
    render(
      <GroupFolderItem
        displayText="Mixed success"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('no-status-folder')).toBeInTheDocument()
  })

  it('shows initializing when Queued and Initializing are present', () => {
    const node = createMockNode({
      [ActionPhase.QUEUED]: 2,
      [ActionPhase.INITIALIZING]: 1,
    })
    render(
      <GroupFolderItem
        displayText="Queued and Initializing"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('initializing-folder')).toBeInTheDocument()
  })

  it('shows queued icon when some children are QUEUED or WAITING_FOR_RESOURCES', () => {
    const node = createMockNode({
      [ActionPhase.QUEUED]: 2,
      [ActionPhase.UNSPECIFIED]: 1,
    })
    render(
      <GroupFolderItem
        displayText="Group B"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('queued-folder')).toBeInTheDocument()
  })

  it('shows Initializing icon when at least one is initializing and none are running or terminal', () => {
    const node = createMockNode({
      [ActionPhase.INITIALIZING]: 1,
      [ActionPhase.UNSPECIFIED]: 2,
    })
    render(
      <GroupFolderItem
        displayText="Group C"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('initializing-folder')).toBeInTheDocument()
  })

  it('shows Running icon when at least one is running and none are terminal', () => {
    const node = createMockNode({
      [ActionPhase.RUNNING]: 1,
      [ActionPhase.UNSPECIFIED]: 1,
    })
    render(
      <GroupFolderItem
        displayText="Group D"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('running-folder')).toBeInTheDocument()
  })

  it('shows Failed icon when any child has FAILED', () => {
    const node = createMockNode({
      [ActionPhase.FAILED]: 1,
      [ActionPhase.RUNNING]: 2,
    })
    render(
      <GroupFolderItem
        displayText="Group E"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('error-folder')).toBeInTheDocument()
  })

  it('shows Timed Out icon if any is timed out and none failed', () => {
    const node = createMockNode({
      [ActionPhase.TIMED_OUT]: 1,
      [ActionPhase.RUNNING]: 2,
    })
    render(
      <GroupFolderItem
        displayText="Group F"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('timed-out-folder')).toBeInTheDocument()
  })

  it('shows Aborted icon if any is aborted and none failed or timed out', () => {
    const node = createMockNode({
      [ActionPhase.ABORTED]: 1,
      [ActionPhase.RUNNING]: 2,
    })
    render(
      <GroupFolderItem
        displayText="Group G"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('aborted-folder')).toBeInTheDocument()
  })

  it('shows succeeded icon if all children succeeded', () => {
    const node = createMockNode({
      [ActionPhase.SUCCEEDED]: 5,
    })
    render(
      <GroupFolderItem
        displayText="Group H"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('success-folder')).toBeInTheDocument()
  })

  it('returns no icon if empty input', () => {
    const node = createMockNode({})
    render(
      <GroupFolderItem
        displayText="Group I"
        isSelected={false}
        maxWidth={200}
        node={node}
      />,
    )
    expect(screen.getByTestId('no-status-folder')).toBeInTheDocument()
  })
})