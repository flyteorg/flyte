import { type Metadata } from 'next'

import { TaskDetailsPage } from '@/components/pages/TaskDetails/Main'

export const metadata: Metadata = {
  title: 'Task Details',
}

/**
 * Task Details Page Route Structure
 *
 * This page handles the dynamic routing for task details with the following segments:
 *
 * Route: /domain/[domain]/project/[project]/tasks/[name]/[[...version]]
 *
 * Segments:
 * - [domain]: Required - The domain name (e.g., "development", "production")
 * - [project]: Required - The project identifier (e.g., "my-project")
 * - [name]: Required - The task name
 * - [version]: Optional - Segment for version information
 *   - when not provided, the page will fetch the latest version of the task
 *   - when provided, the page will fetch the task with the given version
 *
 * Examples:
 * - /domain/dev/project/my-project/tasks/task-name
 * - /domain/prod/project/my-project/tasks/task-name/v1.0.0
 * - /domain/staging/project/my-project/tasks/task-name/latest
 *
 * @see https://nextjs.org/docs/pages/building-your-application/routing/dynamic-routes#optional-catch-all-segments
 */
export default function Home() {
  return <TaskDetailsPage />
}
