## Overview

This document provides a comprehensive guide to the End-to-End (E2E) testing workflow for the `clientsv2` project. It covers how to run tests locally, how they are executed in the Buildkite CI environment, tips for debugging failures, and guidelines for developing new tests.

E2E tests simulate real user scenarios and validate the entire application stack, ensuring that the system works as expected from the user's perspective.

---

## Local Development

### Prerequisites

- node.js
- pmpm installed
- Access to the required environment variables (`HOST`, `TEST_PROJECT`, `TEST_DOMAIN`, `UNION_API_KEY`)

### Setup

1. Clone the repository and navigate to the project directory.
2. Install dependencies:

   ```bash
   pnpm install
   ```

3. Set environment variables locally or add to a .env file at root of clientsv2. While these values specify testing in playground.canary, you can also change these values out for testing in demo.hosted.

   ```bash
   BASE_URL="https://playground.canary.unionai.cloud"
   HOST="playground.canary.unionai.cloud"
   TEST_PROJECT='e2e'
   TEST_DOMAIN="development"
   ADMIN_USERNAME=<username>
   ADMIN_PASSWORD=<password>
   VIEWER_USERNAME=<username>
   VIEWER_PASSWORD=<password>
   UNION_API_KEY=<api_key>
   ```

   Here are some

   ```bash
   BASE_URL="https://demo.hosted.unionai.cloud"
   HOST="demo.hosted.unionai.cloud"
   ADMIN_USERNAME=<username>
   ADMIN_PASSWORD=<password>
   VIEWER_USERNAME=<username>
   VIEWER_PASSWORD=<password>
   TEST_ORG='demo'
   UNION_API_KEY=<api_key>
   ```

4. You must start a development server locally with `pnpm dev` before running tests locally

### Running Tests Locally

1. Ensure the correct environment variables, such as `UNION_API_KEY` have been set.

2. Run the local seed script with `.scripts/e2e-local.sh`.

3. Run the E2E tests using the following command:

```bash
pnpm test:e2e
```

This command uses the configured base URL and environment variables to run tests against your local or specified environment.

---

## Buildkite CI Workflow

The E2E tests are integrated into the Buildkite pipeline to ensure continuous validation on every commit and pull request.

### Workflow Steps

1. Checkout the code.
2. Install dependencies.
3. Set the E2E_ENV environment variable in the Buildkite pipeline settings to specify whether to run on canary or demo.hosted
4. Start development server on port 8080.
5. Run the E2E test suite.
6. Collect and upload test artifacts and reports.

## Writing New Tests

### Test Structure

- Tests are located in the `clientsv2/e2e` directory.
- Use descriptive test names and organize tests by feature or module.
- Follow the existing code style and conventions.

### Best Practices

- Write tests that are deterministic and independent. Tests should avoid relying on data pre-existing in the testing environment, because it can be deleted or removed due to data retention policies at any time.
- Use environment variables to avoid hardcoding URLs and credentials.

### Running Your Tests

Run your new tests locally before pushing:

```bash
yarn e2e path/to/your/test.spec.ts
```

When running locally, we can run in safari and firefox by running `pnpm test:e2e --project=chromium --project=webkit --project=firefox` but this has not been reliable in buildkite, so we only run chromium in CI.

While the tests default to headless mode, you can run `pnpm test:e2e:ui` to watch the runner execute the tests.

---

## Folder Structure

```
clientsv2/
└── e2e/
    ├── constants.ts       # Configuration constants for tests
    ├── tests/             # E2E test files
    ├── helpers/           # Utility and helper functions
    ├── scripts/           # Scripts for seeding the testing environment
    └── README.md          # This documentation
```
