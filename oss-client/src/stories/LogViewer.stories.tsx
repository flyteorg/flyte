import { Meta, StoryObj } from '@storybook/nextjs'
import { LogViewer } from '@/components/LogViewer'
import {
  LogLine,
  LogLineSchema,
} from '@/gen/flyteidl2/logs/dataplane/payload_pb'
import { LogLineOriginator } from '@/gen/flyteidl2/logs/dataplane/payload_pb'
import { RunLogType } from '@/components/pages/RunDetails/types'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
import { create } from '@bufbuild/protobuf'

const meta: Meta<typeof LogViewer> = {
  title: 'components/LogViewer',
  component: LogViewer,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component:
          'A component that displays logs with filtering by source and search functionality.',
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="h-screen">
        <Story />
      </div>
    ),
  ],
}

export default meta
type Story = StoryObj<typeof LogViewer>

const generateLogs = (count: number): LogLine[] => {
  const messages = [
    'Assigned to cluster [union-us-central1]',
    'IngressNotConfigured: Ingress has not yet been reconciled.',
    'App is pending deployment',
    'Remediation task',
    'TrafficNotMigrated: Traffic is not yet migrated to the latest revision.',
    'The legacy algorithm is back online. Status update: Network integrity is in jeopardy.',
    '{"name": "user_space", "levelname": "INFO", "message": "Please open the browser to connect to the running server."}',
  ]

  return Array.from({ length: count }, (_, i) => {
    const timestamp = timestampFromDate(new Date('2024-04-01T05:22:15.008Z'))
    return create(LogLineSchema, {
      timestamp,
      message: messages[i % messages.length],
      originator:
        i % 2 === 0 ? LogLineOriginator.USER : LogLineOriginator.SYSTEM,
    })
  })
}

/**
 * Default view with the standard set of logs
 */
export const Default: Story = {
  args: {
    logs: generateLogs(5),
    logType: RunLogType.RUN,
  },
}

/**
 * Shows how the component handles a large number of logs with scrolling
 */
export const ManyLogs: Story = {
  args: {
    logs: generateLogs(50),
    logType: RunLogType.RUN,
  },
}

/**
 * Shows PyTorch Lightning training logs with distributed training setup
 */
export const TorchLightningLogs: Story = {
  args: {
    logType: RunLogType.RUN,
    logs: [
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:19:50Z')),
        message: 'Starting training session',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:19:56Z')),
        message: 'Using bfloat16 Automatic Mixed Precision (AMP)',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:19:56Z')),
        message: 'GPU available: True (cuda), used: True',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:19:56Z')),
        message: 'TPU available: False, using: 0 TPU cores',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:19:56Z')),
        message: 'HPU available: False, using: 0 HPUs',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:19:56Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 0, MEMBER: 1/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:11Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 5, MEMBER: 6/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:11Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 4, MEMBER: 5/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 6, MEMBER: 7/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 1, MEMBER: 2/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 3, MEMBER: 4/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 2, MEMBER: 3/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message: 'Initializing distributed: GLOBAL_RANK: 7, MEMBER: 8/8',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message: 'distributed_backend=nccl',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:12Z')),
        message:
          'All distributed processes registered. Starting with 8 processes',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:15Z')),
        message: 'GitPython could not be initialized',
        originator: LogLineOriginator.SYSTEM,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:15Z')),
        message:
          'Neptune initialized. Open in the app: https://app.neptune.ai/grantham/windmark/e/WINDMARK-20',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 0 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 2 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 1 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 3 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 5 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 7 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 6 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'LOCAL_RANK: 4 - CUDA_VISIBLE_DEVICES: [0,1,2,3,4,5,6,7]',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '┏━━━┳━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━┳━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━┓',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '┃   ┃                        ┃          ┃        ┃       ┃          ┃      Out ┃',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '┃   ┃ Name                   ┃ Type     ┃ Params ┃ Mode  ┃ In sizes ┃    sizes ┃',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '┡━━━╇━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━╇━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━┩',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 0 │ modular_field_embedder │ Modular… │ 34.8 K │ train │    [108] │   [[108, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │          │   96, 6, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │          │     64], │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │          │ [108, 0, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │          │    384]] │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 1 │ dynamic_field_encoder  │ Dynamic… │  281 K │ train │    [108, │    [108, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │   96, 6, │ 96, 384] │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │      64] │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 2 │ event_encoder          │ EventEn… │  8.7 M │ train │   [[108, │   [[108, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │      96, │    384], │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    384], │    [108, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │ [108, 0, │      96, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    384]] │    384], │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │          │ [108, 0, │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │          │    384]] │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 3 │ static_field_decoder   │ StaticF… │      0 │ train │   [[108, │    [108] │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │ 0, 384], │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    [108, │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    384]] │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 4 │ event_decoder          │ EventDe… │  7.1 M │ train │   [[108, │    [108] │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │      96, │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    384], │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    [108, │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │    384]] │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 5 │ decision_head          │ Decisio… │  116 K │ train │    [108, │ [108, 2] │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│   │                        │          │        │       │     384] │          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '│ 6 │ metrics                │ ModuleD… │      0 │ train │        ? │        ? │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message:
          '└───┴────────────────────────┴──────────┴────────┴───────┴──────────┴──────────┘',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'Trainable params: 16.1 M',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'Non-trainable params: 116 K',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'Total params: 16.3 M',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'Total estimated model params size (MB): 65',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'Modules in train mode: 131',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:20:18Z')),
        message: 'Modules in eval mode: 0',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:21:43Z')),
        message:
          "Grad strides do not match bucket view strides. This may indicate grad was not created according to the gradient layout contract, or that the param's strides changed since DDP was constructed. This is not an error, but may impair performance.",
        originator: LogLineOriginator.SYSTEM,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:21:43Z')),
        message: 'grad.sizes() = [1, 1, 384], strides() = [37248, 384, 1]',
        originator: LogLineOriginator.SYSTEM,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:21:43Z')),
        message: 'bucket_view.sizes() = [1, 1, 384], strides() = [384, 384, 1]',
        originator: LogLineOriginator.SYSTEM,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:23:04Z')),
        message:
          'Error occurred during asynchronous operation processing: X-coordinates (step) must be strictly increasing for series attribute: training/epoch. Invalid point: 63.0',
        originator: LogLineOriginator.SYSTEM,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:26:10Z')),
        message: 'Trainer.fit stopped: max_epochs=4 reached.',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:26:13Z')),
        message:
          'Epoch 3/3 ━━━━━━━━━━━━━━━━━ 64/64 0:00:42 • 0:00:00 1.82it/s v_num: K-20',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:26:13Z')),
        message: 'pretrain-total-va…',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:26:13Z')),
        message: '18.863',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '┃             Test metric              ┃             DataLoader 0             ┃',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│     pretrain-dynamic-test/amount     │          4.131650447845459           │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│ pretrain-dynamic-test/merchant_city  │          4.368139266967773           │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│ pretrain-dynamic-test/merchant_name  │          4.567246913909912           │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│ pretrain-dynamic-test/merchant_state │          2.4917099475860596          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│   pretrain-dynamic-test/timestamp    │          3.2738277912139893          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│    pretrain-dynamic-test/use_chip    │          0.4375220835208893          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│     pretrain-total-test/dynamic      │          19.270097732543945          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '│       pretrain-total-test/loss       │          19.270097732543945          │',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:18Z')),
        message:
          '└──────────────────────────────────────┴──────────────────────────────────────┘',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:19Z')),
        message:
          'Testing ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 129/-- 0:00:21 • -:--:-- 6.62it/s',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:21Z')),
        message: 'All 4 operations synced, thanks for waiting!',
        originator: LogLineOriginator.USER,
      }),
      create(LogLineSchema, {
        timestamp: timestampFromDate(new Date('2025-04-09T16:27:21Z')),
        message:
          'Explore the metadata in the Neptune app: https://app.neptune.ai/grantham/windmark/e/WINDMARK-20/metadata',
        originator: LogLineOriginator.USER,
      }),
    ],
  },
}

/**
 * Shows how the component handles a single source type
 */
export const SingleSource: Story = {
  args: {
    logs: generateLogs(5).map((log) =>
      create(LogLineSchema, {
        ...log,
        originator: LogLineOriginator.USER,
      }),
    ),
    logType: RunLogType.RUN,
  },
}

/**
 * Shows how the component handles error states
 */
export const ErrorLogs: Story = {
  args: {
    logs: generateLogs(5).map((log) =>
      create(LogLineSchema, {
        ...log,
        originator: LogLineOriginator.SYSTEM,
      }),
    ),
    logType: RunLogType.RUN,
  },
}

/**
 * Shows how the component handles warning states
 */
export const WarningLogs: Story = {
  args: {
    logs: generateLogs(5).map((log) =>
      create(LogLineSchema, {
        ...log,
        originator: LogLineOriginator.SYSTEM,
      }),
    ),
    logType: RunLogType.RUN,
  },
}

/**
 * Shows how the component handles empty state
 */
export const NoLogs: Story = {
  args: {
    logs: [],
  },
}
