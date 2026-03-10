import { ActionTimeline } from '@/components/pages/RunDetails/ActionPhaseTimeline'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { PhaseTransition } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import { Meta, StoryObj } from '@storybook/nextjs'
import React from 'react'

type Story = StoryObj<typeof ActionTimeline>

const meta: Meta<typeof ActionTimeline> = {
  title: 'Components/ActionTimeline',
  component: ActionTimeline,
  args: {
    taskType: 'python',
  },
}

const makeTimestamp = (seconds: number, nanos: number): Timestamp => ({
  seconds: BigInt(seconds),
  nanos: nanos,
  $typeName: 'google.protobuf.Timestamp',
})

export default meta

const ts = (millis: number) => {
  const seconds = Math.floor(millis / 1000)
  const nanos = (millis % 1000) * 1_000_000
  return makeTimestamp(seconds, nanos)
}

const makeTransition = (phase: ActionPhase, start: number, end?: number) =>
  ({
    phase,
    startTime: ts(start),
    endTime: end !== undefined ? ts(end) : undefined,
  }) as PhaseTransition

export const NotStarted: Story = {
  args: {
    phaseTransitions: [makeTransition(ActionPhase.UNSPECIFIED, 100, 2000)],
  },
}

export const Queued: Story = {
  args: { phaseTransitions: [makeTransition(ActionPhase.QUEUED, 100, 2000)] },
}

export const Running: Story = {
  args: {
    phaseTransitions: [
      makeTransition(ActionPhase.QUEUED, 100, 10000),
      makeTransition(ActionPhase.INITIALIZING, 10000, 30000),
      makeTransition(ActionPhase.RUNNING, 30000, 50000),
    ],
  },
}

export const Succeeded: Story = {
  args: {
    phaseTransitions: [
      makeTransition(ActionPhase.QUEUED, 0, 2000),
      makeTransition(ActionPhase.INITIALIZING, 2000, 3000),
      makeTransition(ActionPhase.RUNNING, 3000, 5000),
      makeTransition(ActionPhase.SUCCEEDED, 5000),
    ],
  },
}

export const FailedQuickly: Story = {
  args: {
    phaseTransitions: [
      makeTransition(ActionPhase.QUEUED, 0, 10),
      makeTransition(ActionPhase.INITIALIZING, 10, 20),
      makeTransition(ActionPhase.RUNNING, 20, 30),
      makeTransition(ActionPhase.FAILED, 30),
    ],
  },
}

export const TimedOut: Story = {
  args: {
    phaseTransitions: [
      makeTransition(ActionPhase.QUEUED, 0, 10),
      makeTransition(ActionPhase.INITIALIZING, 10, 20),
      makeTransition(ActionPhase.RUNNING, 20, 10000000),
      makeTransition(ActionPhase.TIMED_OUT, 10000000),
    ],
  },
}

export const Aborted: Story = {
  args: {
    phaseTransitions: [
      makeTransition(ActionPhase.QUEUED, 10, 20),
      makeTransition(ActionPhase.INITIALIZING, 20, 30),
      makeTransition(ActionPhase.RUNNING, 30, 100),
      makeTransition(ActionPhase.ABORTED, 150),
    ],
  },
}

const AnimatedLifecycleComponent: React.FC = () => {
  const [phaseTransitions, setPhaseTransitions] = React.useState<
    PhaseTransition[]
  >([
    makeTransition(ActionPhase.QUEUED, 0), // initially only QUEUED
  ])

  React.useEffect(() => {
    const now = Date.now()

    const steps = [
      { delay: 1000, phase: ActionPhase.QUEUED, start: now, end: now + 2000 },
      {
        delay: 2000,
        phase: ActionPhase.INITIALIZING,
        start: now + 2000,
        end: now + 3000,
      },
      { delay: 3000, phase: ActionPhase.RUNNING, start: now + 3000 },
      {
        delay: 6000,
        phase: ActionPhase.SUCCEEDED,
        start: now + 6000,
        end: now + 4000,
      },
    ]

    steps.forEach(({ delay, phase, start, end }) => {
      setTimeout(() => {
        setPhaseTransitions((prev) => [
          ...prev.filter((p) => p.phase !== phase),
          makeTransition(phase, start, end),
        ])
      }, delay)
    })
  }, [])

  return (
    <ActionTimeline taskType="python" phaseTransitions={phaseTransitions} />
  )
}

export const AnimatedLifecycle: Story = {
  render: () => <AnimatedLifecycleComponent />,
}
