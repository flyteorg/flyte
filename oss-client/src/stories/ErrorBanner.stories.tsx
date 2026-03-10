import type { Meta, StoryObj } from '@storybook/nextjs'
import { ErrorBanner } from '@/components/ErrorBanner'
import {
  AbortInfo,
  ErrorInfo,
  ErrorInfo_Kind,
} from '@/gen/flyteidl2/workflow/run_definition_pb'

const meta = {
  title: 'components/ErrorBanner',
  component: ErrorBanner,
  parameters: {
    layout: 'fullscreen',
    design: {
      type: 'figma',
      url: 'https://www.figma.com/design/8srmhAcZjvRfCgSGDvwbfZ/Workflows-V2?node-id=846-21691&t=Q1uE3LH1x8SSRs81-0',
    },
  },
  decorators: [
    (Story) => (
      <div className="min-w-[320px] p-4">
        <Story />
      </div>
    ),
  ],
  tags: ['autodocs'],
} satisfies Meta<typeof ErrorBanner>

export default meta
type Story = StoryObj<typeof meta>

export const UserError: Story = {
  args: {
    result: {
      case: 'errorInfo',
      value: {
        message:
          'Invalid input parameter: "max_retries" must be a positive integer',
        kind: ErrorInfo_Kind.USER,
      } as ErrorInfo,
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Displays a user error with a red badge indicating user input validation failure.',
      },
    },
  },
}

export const SystemError: Story = {
  args: {
    result: {
      case: 'errorInfo',
      value: {
        message:
          'Internal server error: Failed to connect to database\nCaused by: Connection timeout after 30s',
        kind: ErrorInfo_Kind.SYSTEM,
      } as ErrorInfo,
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Shows a system error with details about internal service failures.',
      },
    },
  },
}

export const UnspecifiedError: Story = {
  args: {
    result: {
      case: 'errorInfo',
      value: {
        message: 'An unexpected error occurred during execution',
        kind: ErrorInfo_Kind.UNSPECIFIED,
      } as ErrorInfo,
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Represents an unspecified error when the error type cannot be determined.',
      },
    },
  },
}

export const WithStackTrace: Story = {
  args: {
    result: {
      case: 'errorInfo',
      value: {
        message: `[1/1] currentAttempt done. Last Error: USER::
[f7f8f879b002690a3000-n0-0] terminated with exit code (1). Reason [Error]. Message:
lyteValueException: USER:ValueError: error=Value error!  Received: s3://union-oc-production-demo/flytesnacks/development/WTYDZNJ7TXE4EKAFDUWVUJUSCA======/faste4bf36a3162216450ab5134001005b5e.tar.gz. File not found

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/usr/local/bin/pyflyte-fast-execute", line 8, in <module>
    sys.exit(fast_execute_task_cmd())
  File "/usr/local/lib/python3.9/site-packages/click/core.py", line 1157, in __call__
    return self.main(*args, **kwargs)
  File "/usr/local/lib/python3.9/site-packages/click/core.py", line 1078, in main
    rv = self.invoke(ctx)
  File "/usr/local/lib/python3.9/site-packages/click/core.py", line 1434, in invoke
    return ctx.invoke(self.callback, **ctx.params)
  File "/usr/local/lib/python3.9/site-packages/click/core.py", line 783, in invoke
    return __callback(*args, **kwargs)
  File "/usr/local/lib/python3.9/site-packages/flytekit/bin/entrypoint.py", line 524, in fast_execute_task_cmd
    _download_distribution(additional_distribution, dest_dir)
  File "/usr/local/lib/python3.9/site-packages/flytekit/core/utils.py", line 309, in wrapper
    return func(*args, **kwargs)
  File "/usr/local/lib/python3.9/site-packages/flytekit/tools/fast_registration.py", line 119, in download_distribution
    FlyteContextManager.current_context().file_access.get_data(additional_distribution, os.path.join(destination, ""))
  File "/usr/local/lib/python3.9/site-packages/flytekit/core/data_persistence.py", line 511, in get_data
    raise FlyteAssertion(
flytekit.exceptions.user.FlyteAssertion: USER:AssertionError: error=Failed to get data from s3://union-oc-production-demo/flytesnacks/development/WTYDZNJ7TXE4EKAFDUWVUJUSCA======/faste4bf36a3162216450ab5134001005b5e.tar.gz to /root/ (recursive=False).

Original exception: USER:ValueError: error=Value error!  Received: s3://union-oc-production-demo/flytesnacks/development/WTYDZNJ7TXE4EKAFDUWVUJUSCA======/faste4bf36a3162216450ab5134001005b5e.tar.gz. File not found
.`,
        kind: ErrorInfo_Kind.SYSTEM,
      } as ErrorInfo,
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Demonstrates how the banner handles long stack traces with expandable content.',
      },
    },
  },
}

export const LongMessage: Story = {
  args: {
    result: {
      case: 'errorInfo',
      value: {
        message:
          'Multiple validation errors occurred:\n1. Parameter "timeout" must be between 1 and 3600 seconds\n2. Required field "retry_policy" is missing\n3. Invalid credentials provided for external service',
        kind: ErrorInfo_Kind.USER,
      } as ErrorInfo,
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Shows how multiple validation errors are displayed in a structured format.',
      },
    },
  },
}

export const Aborted: Story = {
  args: {
    result: {
      case: 'abortInfo',
      value: {
        reason: 'User requested cancellation',
      } as AbortInfo,
    },
  },
  parameters: {
    docs: {
      description: {
        story:
          'Displays the abort state with an orange badge and keeps the abort message visible even when expanded.',
      },
    },
  },
}
