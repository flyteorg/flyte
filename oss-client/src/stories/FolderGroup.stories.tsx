import React from 'react'
import { Meta, StoryObj } from '@storybook/nextjs'
import { FolderGroup } from '@/components/FolderGroup'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

const meta: Meta<typeof FolderGroup> = {
  title: 'Icons/FolderGroup',
  component: FolderGroup,
  parameters: {
    docs: {
      description: {
        component:
          'Displays a folder icon based on the phase. Each phase renders a different SVG variant for visual distinction.',
      },
    },
  },
  argTypes: {
    phase: {
      control: { type: 'select' },
      options: Object.keys(ActionPhase).map(
        (key) => ActionPhase[key as keyof typeof ActionPhase],
      ),
      mapping: ActionPhase,
    },
  },
}

export default meta

type Story = StoryObj<typeof FolderGroup>

export const Default: Story = {
  args: {
    phase: ActionPhase.UNSPECIFIED,
    props: {
      width: 40,
      height: 40,
    },
  },
  parameters: {
    docs: {
      description: {
        story: 'Default folder icon based on the UNSPECIFIED phase.',
      },
    },
  },
}

export const AllPhases: Story = {
  render: (_args) => (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 20 }}>
      {Object.values(ActionPhase)
        .filter((v) => typeof v === 'number')
        .map((p) => (
          <div
            key={p as number}
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              width: 80,
            }}
          >
            <FolderGroup phase={p as ActionPhase} props={{ width: 40, height: 40 }} />
            <div style={{ fontSize: 12, marginTop: 4, textAlign: 'center' }}>
              {ActionPhase[p as number]}
            </div>
          </div>
        ))}
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Displays all phase icons for visual comparison.',
      },
    },
  },
}